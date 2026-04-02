export namespace model {
	
	export class QRCodeData {
	    url: string;
	    qrcode_key: string;
	
	    static createFrom(source: any = {}) {
	        return new QRCodeData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.qrcode_key = source["qrcode_key"];
	    }
	}

}

