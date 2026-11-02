package netclient

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"

	"testing"
	"time"

	"github.com/imroc/req/v3"
)

func TestProxyValidation(t *testing.T) {
	for _, input := range []string{"", " http://127.0.0.1:7890/ ", "https://user:pass@proxy.test:443", "socks5://127.0.0.1:1080", "socks5h://proxy.test:1080"} {
		if _, err := NormalizeProxy(input); err != nil {
			t.Errorf("rejected valid proxy: %v", err)
		}
	}
	for _, input := range []string{"127.0.0.1:7890", "ftp://proxy.test:80", "http://:7890", "http://proxy.test:65536", "http://proxy.test:0", "http://proxy.test/path", "http://proxy.test?x=1", "socks5://proxy.test", "http://%", "http://proxy.test\n" + "bad"} {
		if _, err := NormalizeProxy(input); err == nil {
			t.Error("accepted invalid proxy address")
		}
	}
}

// Exercise real req proxy routing through an in-memory connection; no listening
// port or external network is needed, including for the SOCKS5 handshake.
func TestHTTPAndSOCKSProxyRouting(t *testing.T) {
	for _, scheme := range []string{"http", "socks5", "socks5h"} {
		t.Run(scheme, func(t *testing.T) {
			client := New(scheme+"://proxy.test:1080", 2*time.Second, nil)
			result := make(chan error, 1)
			client.SetDial(func(ctx context.Context, network, address string) (net.Conn, error) {
				if address != "proxy.test:1080" {
					return nil, fmt.Errorf("dialed target directly: %s", address)
				}
				local, remote := net.Pipe()
				go func() {
					defer remote.Close()
					remote.SetDeadline(time.Now().Add(2 * time.Second))
					if scheme != "http" {
						greeting := make([]byte, 3)
						if _, err := io.ReadFull(remote, greeting); err != nil {
							result <- err
							return
						}
						if greeting[0] != 5 || greeting[1] != 1 || greeting[2] != 0 {
							result <- fmt.Errorf("bad SOCKS greeting")
							return
						}
						if _, err := remote.Write([]byte{5, 0}); err != nil {
							result <- err
							return
						}
						header := make([]byte, 5)
						if _, err := io.ReadFull(remote, header); err != nil {
							result <- err
							return
						}
						if header[0] != 5 || header[1] != 1 || header[3] != 3 {
							result <- fmt.Errorf("bad SOCKS connect")
							return
						}
						target := make([]byte, int(header[4])+2)
						if _, err := io.ReadFull(remote, target); err != nil {
							result <- err
							return
						}
						if string(target[:len(target)-2]) != "media.test" {
							result <- fmt.Errorf("SOCKS target mismatch")
							return
						}
						if _, err := remote.Write([]byte{5, 0, 0, 1, 127, 0, 0, 1, 0, 80}); err != nil {
							result <- err
							return
						}
					}
					request, err := http.ReadRequest(bufio.NewReader(remote))
					if err != nil {
						result <- err
						return
					}
					defer request.Body.Close()
					if request.Host != "media.test" || request.URL.Path != "/video" {
						result <- fmt.Errorf("wrong target")
						return
					}
					_, err = io.WriteString(remote, "HTTP/1.1 200 OK\r\nContent-Length: 4\r\nConnection: close\r\n\r\ndata")
					result <- err
				}()
				return local, nil
			})
			response, err := client.R().Get("http://media.test/video")
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			data, err := io.ReadAll(response.Body)
			if err != nil || string(data) != "data" {
				t.Fatalf("bad streaming response %q: %v", data, err)
			}
			if err := <-result; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestHTTPSDownloadThroughAuthenticatedConnectProxy(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "media.test"}, DNSNames: []string{"media.test"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	client := New("http://user:pass@proxy.test:1080", 2*time.Second, nil)
	client.GetTLSClientConfig().RootCAs = pool
	result := make(chan error, 1)
	client.SetDial(func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "proxy.test:1080" {
			return nil, fmt.Errorf("bypassed proxy")
		}
		local, remote := net.Pipe()
		go func() {
			defer remote.Close()
			remote.SetDeadline(time.Now().Add(2 * time.Second))
			connect, err := http.ReadRequest(bufio.NewReader(remote))
			if err != nil {
				result <- err
				return
			}
			if connect.Method != "CONNECT" || connect.Host != "media.test:443" || connect.Header.Get("Proxy-Authorization") != "Basic dXNlcjpwYXNz" {
				result <- fmt.Errorf("invalid CONNECT target or proxy authentication")
				return
			}
			if _, err = io.WriteString(remote, "HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
				result <- err
				return
			}
			tunnel := tls.Server(remote, &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}})
			// Close the pipe directly; TLS close_notify could block after the client
			// has consumed the complete body on an in-memory unbuffered connection.
			request, err := http.ReadRequest(bufio.NewReader(tunnel))
			if err != nil {
				result <- err
				return
			}
			defer request.Body.Close()
			if request.Header.Get("Proxy-Authorization") != "" {
				result <- fmt.Errorf("proxy credentials leaked to origin")
				return
			}
			_, err = io.WriteString(tunnel, "HTTP/1.1 200 OK\r\nContent-Length: 4\r\nConnection: close\r\n\r\ndata")
			result <- err
		}()
		return local, nil
	})
	response, err := client.R().Get("https://media.test/video")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil || string(data) != "data" {
		t.Fatalf("invalid HTTPS body: %v", err)
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
}

func TestStreamBodyRemainsCancelable(t *testing.T) {
	client := New("", 0, nil).SetProxy(nil)
	client.SetDial(func(ctx context.Context, network, address string) (net.Conn, error) {
		local, remote := net.Pipe()
		go func() {
			defer remote.Close()
			remote.SetDeadline(time.Now().Add(2 * time.Second))
			request, err := http.ReadRequest(bufio.NewReader(remote))
			if err != nil {
				return
			}
			request.Body.Close()
			io.WriteString(remote, "HTTP/1.1 200 OK\r\nContent-Length: 8\r\n\r\ndata")
			// Wait for cancellation to close the connection before sending the rest.
			var b [1]byte
			remote.Read(b[:])
		}()
		return local, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response, err := client.R().SetContext(ctx).Get("http://media.test/video")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var first [4]byte
	if _, err := io.ReadFull(response.Body, first[:]); err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err := io.ReadAll(response.Body); err == nil {
		t.Fatal("canceled body read succeeded")
	}
}

func TestExplicitProxySwitchIgnoresEnvironmentWhenDisabled(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://environment.test:8080")
	t.Setenv("HTTPS_PROXY", "http://environment.test:8080")
	for _, enabled := range []bool{false, true} {
		client := NewWithProxy(DefaultProxy, enabled, time.Second, nil)
		// Media downloads clone the parse client's transport.
		for _, transport := range []*req.Transport{client.GetTransport(), client.Clone().GetTransport()} {
			if !enabled {
				if transport.Proxy != nil {
					t.Fatal("disabled proxy still uses environment settings")
				}
				continue
			}
			target, _ := url.Parse("https://www.youtube.com/watch?v=fixture")
			proxy, err := transport.Proxy(&http.Request{URL: target})
			if err != nil || proxy == nil || proxy.String() != DefaultProxy {
				t.Fatalf("wrong enabled proxy: %v %v", proxy, err)
			}
		}
	}
}
