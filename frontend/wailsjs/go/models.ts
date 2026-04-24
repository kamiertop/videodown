export namespace api {
	
	export class DashDownloadResult {
	    bvid: string;
	    cid: number;
	    title: string;
	    path: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new DashDownloadResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bvid = source["bvid"];
	        this.cid = source["cid"];
	        this.title = source["title"];
	        this.path = source["path"];
	        this.error = source["error"];
	    }
	}
	export class DashDownloadBatchResult {
	    results: DashDownloadResult[];
	    success: number;
	    failed: number;
	
	    static createFrom(source: any = {}) {
	        return new DashDownloadBatchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.results = this.convertValues(source["results"], DashDownloadResult);
	        this.success = source["success"];
	        this.failed = source["failed"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class DashDownloadTask {
	    sourceName: string;
	    sourceKind: string;
	    upperName: string;
	    bvid: string;
	    cid: number;
	    title: string;
	    cover: string;
	    duration: number;
	    play: number;
	    danmaku: number;
	    pubtime: number;
	    videoURL: string;
	    audioURL: string;
	
	    static createFrom(source: any = {}) {
	        return new DashDownloadTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceName = source["sourceName"];
	        this.sourceKind = source["sourceKind"];
	        this.upperName = source["upperName"];
	        this.bvid = source["bvid"];
	        this.cid = source["cid"];
	        this.title = source["title"];
	        this.cover = source["cover"];
	        this.duration = source["duration"];
	        this.play = source["play"];
	        this.danmaku = source["danmaku"];
	        this.pubtime = source["pubtime"];
	        this.videoURL = source["videoURL"];
	        this.audioURL = source["audioURL"];
	    }
	}
	export class DownloadHistoryItem {
	    bvid: string;
	    title: string;
	    cover: string;
	    duration: number;
	    upperName: string;
	    play: number;
	    danmaku: number;
	    pubtime: number;
	    sourceName: string;
	    sourceKind: string;
	    path: string;
	    downloaded: string;
	
	    static createFrom(source: any = {}) {
	        return new DownloadHistoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bvid = source["bvid"];
	        this.title = source["title"];
	        this.cover = source["cover"];
	        this.duration = source["duration"];
	        this.upperName = source["upperName"];
	        this.play = source["play"];
	        this.danmaku = source["danmaku"];
	        this.pubtime = source["pubtime"];
	        this.sourceName = source["sourceName"];
	        this.sourceKind = source["sourceKind"];
	        this.path = source["path"];
	        this.downloaded = source["downloaded"];
	    }
	}

}

export namespace badger {
	
	export class KVLoader {
	
	
	    static createFrom(source: any = {}) {
	        return new KVLoader(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class LevelInfo {
	    Level: number;
	    NumTables: number;
	    Size: number;
	    TargetSize: number;
	    TargetFileSize: number;
	    IsBaseLevel: boolean;
	    Score: number;
	    Adjusted: number;
	    StaleDatSize: number;
	
	    static createFrom(source: any = {}) {
	        return new LevelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Level = source["Level"];
	        this.NumTables = source["NumTables"];
	        this.Size = source["Size"];
	        this.TargetSize = source["TargetSize"];
	        this.TargetFileSize = source["TargetFileSize"];
	        this.IsBaseLevel = source["IsBaseLevel"];
	        this.Score = source["Score"];
	        this.Adjusted = source["Adjusted"];
	        this.StaleDatSize = source["StaleDatSize"];
	    }
	}
	export class MergeOperator {
	
	
	    static createFrom(source: any = {}) {
	        return new MergeOperator(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class Options {
	    Dir: string;
	    ValueDir: string;
	    SyncWrites: boolean;
	    NumVersionsToKeep: number;
	    ReadOnly: boolean;
	    Logger: any;
	    Compression: number;
	    InMemory: boolean;
	    MetricsEnabled: boolean;
	    NumGoroutines: number;
	    MemTableSize: number;
	    BaseTableSize: number;
	    BaseLevelSize: number;
	    LevelSizeMultiplier: number;
	    TableSizeMultiplier: number;
	    MaxLevels: number;
	    VLogPercentile: number;
	    ValueThreshold: number;
	    NumMemtables: number;
	    BlockSize: number;
	    BloomFalsePositive: number;
	    BlockCacheSize: number;
	    IndexCacheSize: number;
	    NumLevelZeroTables: number;
	    NumLevelZeroTablesStall: number;
	    ValueLogFileSize: number;
	    ValueLogMaxEntries: number;
	    NumCompactors: number;
	    CompactL0OnClose: boolean;
	    LmaxCompaction: boolean;
	    ZSTDCompressionLevel: number;
	    VerifyValueChecksum: boolean;
	    EncryptionKey: number[];
	    EncryptionKeyRotationDuration: number;
	    BypassLockGuard: boolean;
	    ChecksumVerificationMode: number;
	    DetectConflicts: boolean;
	    NamespaceOffset: number;
	    ExternalMagicVersion: number;
	
	    static createFrom(source: any = {}) {
	        return new Options(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Dir = source["Dir"];
	        this.ValueDir = source["ValueDir"];
	        this.SyncWrites = source["SyncWrites"];
	        this.NumVersionsToKeep = source["NumVersionsToKeep"];
	        this.ReadOnly = source["ReadOnly"];
	        this.Logger = source["Logger"];
	        this.Compression = source["Compression"];
	        this.InMemory = source["InMemory"];
	        this.MetricsEnabled = source["MetricsEnabled"];
	        this.NumGoroutines = source["NumGoroutines"];
	        this.MemTableSize = source["MemTableSize"];
	        this.BaseTableSize = source["BaseTableSize"];
	        this.BaseLevelSize = source["BaseLevelSize"];
	        this.LevelSizeMultiplier = source["LevelSizeMultiplier"];
	        this.TableSizeMultiplier = source["TableSizeMultiplier"];
	        this.MaxLevels = source["MaxLevels"];
	        this.VLogPercentile = source["VLogPercentile"];
	        this.ValueThreshold = source["ValueThreshold"];
	        this.NumMemtables = source["NumMemtables"];
	        this.BlockSize = source["BlockSize"];
	        this.BloomFalsePositive = source["BloomFalsePositive"];
	        this.BlockCacheSize = source["BlockCacheSize"];
	        this.IndexCacheSize = source["IndexCacheSize"];
	        this.NumLevelZeroTables = source["NumLevelZeroTables"];
	        this.NumLevelZeroTablesStall = source["NumLevelZeroTablesStall"];
	        this.ValueLogFileSize = source["ValueLogFileSize"];
	        this.ValueLogMaxEntries = source["ValueLogMaxEntries"];
	        this.NumCompactors = source["NumCompactors"];
	        this.CompactL0OnClose = source["CompactL0OnClose"];
	        this.LmaxCompaction = source["LmaxCompaction"];
	        this.ZSTDCompressionLevel = source["ZSTDCompressionLevel"];
	        this.VerifyValueChecksum = source["VerifyValueChecksum"];
	        this.EncryptionKey = source["EncryptionKey"];
	        this.EncryptionKeyRotationDuration = source["EncryptionKeyRotationDuration"];
	        this.BypassLockGuard = source["BypassLockGuard"];
	        this.ChecksumVerificationMode = source["ChecksumVerificationMode"];
	        this.DetectConflicts = source["DetectConflicts"];
	        this.NamespaceOffset = source["NamespaceOffset"];
	        this.ExternalMagicVersion = source["ExternalMagicVersion"];
	    }
	}
	export class Sequence {
	
	
	    static createFrom(source: any = {}) {
	        return new Sequence(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class Stream {
	    Prefix: number[];
	    NumGo: number;
	    LogPrefix: string;
	    MaxSize: number;
	    UseKeyToListWithThreadId: boolean;
	    SinceTs: number;
	
	    static createFrom(source: any = {}) {
	        return new Stream(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Prefix = source["Prefix"];
	        this.NumGo = source["NumGo"];
	        this.LogPrefix = source["LogPrefix"];
	        this.MaxSize = source["MaxSize"];
	        this.UseKeyToListWithThreadId = source["UseKeyToListWithThreadId"];
	        this.SinceTs = source["SinceTs"];
	    }
	}
	export class StreamWriter {
	
	
	    static createFrom(source: any = {}) {
	        return new StreamWriter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class TableInfo {
	    ID: number;
	    Level: number;
	    Left: number[];
	    Right: number[];
	    KeyCount: number;
	    OnDiskSize: number;
	    StaleDataSize: number;
	    UncompressedSize: number;
	    MaxVersion: number;
	    IndexSz: number;
	    BloomFilterSize: number;
	
	    static createFrom(source: any = {}) {
	        return new TableInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Level = source["Level"];
	        this.Left = source["Left"];
	        this.Right = source["Right"];
	        this.KeyCount = source["KeyCount"];
	        this.OnDiskSize = source["OnDiskSize"];
	        this.StaleDataSize = source["StaleDataSize"];
	        this.UncompressedSize = source["UncompressedSize"];
	        this.MaxVersion = source["MaxVersion"];
	        this.IndexSz = source["IndexSz"];
	        this.BloomFilterSize = source["BloomFilterSize"];
	    }
	}
	export class Txn {
	
	
	    static createFrom(source: any = {}) {
	        return new Txn(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class WriteBatch {
	
	
	    static createFrom(source: any = {}) {
	        return new WriteBatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class keyRange {
	
	
	    static createFrom(source: any = {}) {
	        return new keyRange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace model {
	
	export class Dimension {
	    width: number;
	    height: number;
	    rotate: number;
	
	    static createFrom(source: any = {}) {
	        return new Dimension(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.width = source["width"];
	        this.height = source["height"];
	        this.rotate = source["rotate"];
	    }
	}
	export class VideoStat {
	    aid: number;
	    view: number;
	    danmaku: number;
	    reply: number;
	    fav: number;
	    coin: number;
	    share: number;
	    now_rank: number;
	    his_rank: number;
	    like: number;
	    dislike: number;
	    evaluation: string;
	    argue_msg: string;
	    vt: number;
	    vv: number;
	
	    static createFrom(source: any = {}) {
	        return new VideoStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.aid = source["aid"];
	        this.view = source["view"];
	        this.danmaku = source["danmaku"];
	        this.reply = source["reply"];
	        this.fav = source["fav"];
	        this.coin = source["coin"];
	        this.share = source["share"];
	        this.now_rank = source["now_rank"];
	        this.his_rank = source["his_rank"];
	        this.like = source["like"];
	        this.dislike = source["dislike"];
	        this.evaluation = source["evaluation"];
	        this.argue_msg = source["argue_msg"];
	        this.vt = source["vt"];
	        this.vv = source["vv"];
	    }
	}
	export class VideoOwner {
	    mid: number;
	    name: string;
	    face: string;
	
	    static createFrom(source: any = {}) {
	        return new VideoOwner(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mid = source["mid"];
	        this.name = source["name"];
	        this.face = source["face"];
	    }
	}
	export class ArcRights {
	    bp: number;
	    elec: number;
	    download: number;
	    movie: number;
	    pay: number;
	    hd5: number;
	    no_reprint: number;
	    autoplay: number;
	    ugc_pay: number;
	    is_cooperation: number;
	    ugc_pay_preview: number;
	    arc_pay: number;
	    free_watch: number;
	
	    static createFrom(source: any = {}) {
	        return new ArcRights(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bp = source["bp"];
	        this.elec = source["elec"];
	        this.download = source["download"];
	        this.movie = source["movie"];
	        this.pay = source["pay"];
	        this.hd5 = source["hd5"];
	        this.no_reprint = source["no_reprint"];
	        this.autoplay = source["autoplay"];
	        this.ugc_pay = source["ugc_pay"];
	        this.is_cooperation = source["is_cooperation"];
	        this.ugc_pay_preview = source["ugc_pay_preview"];
	        this.arc_pay = source["arc_pay"];
	        this.free_watch = source["free_watch"];
	    }
	}
	export class Arc {
	    aid: number;
	    videos: number;
	    type_id: number;
	    type_name: string;
	    copyright: number;
	    pic: string;
	    title: string;
	    pubdate: number;
	    ctime: number;
	    desc: string;
	    state: number;
	    duration: number;
	    rights: ArcRights;
	    author: VideoOwner;
	    stat: VideoStat;
	    dynamic: string;
	    dimension: Dimension;
	    desc_v2: any;
	    is_chargeable_season: boolean;
	    is_blooper: boolean;
	    enable_vt: number;
	    vt_display: string;
	    type_id_v2: number;
	    type_name_v2: string;
	    is_lesson_video: number;
	
	    static createFrom(source: any = {}) {
	        return new Arc(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.aid = source["aid"];
	        this.videos = source["videos"];
	        this.type_id = source["type_id"];
	        this.type_name = source["type_name"];
	        this.copyright = source["copyright"];
	        this.pic = source["pic"];
	        this.title = source["title"];
	        this.pubdate = source["pubdate"];
	        this.ctime = source["ctime"];
	        this.desc = source["desc"];
	        this.state = source["state"];
	        this.duration = source["duration"];
	        this.rights = this.convertValues(source["rights"], ArcRights);
	        this.author = this.convertValues(source["author"], VideoOwner);
	        this.stat = this.convertValues(source["stat"], VideoStat);
	        this.dynamic = source["dynamic"];
	        this.dimension = this.convertValues(source["dimension"], Dimension);
	        this.desc_v2 = source["desc_v2"];
	        this.is_chargeable_season = source["is_chargeable_season"];
	        this.is_blooper = source["is_blooper"];
	        this.enable_vt = source["enable_vt"];
	        this.vt_display = source["vt_display"];
	        this.type_id_v2 = source["type_id_v2"];
	        this.type_name_v2 = source["type_name_v2"];
	        this.is_lesson_video = source["is_lesson_video"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class SegmentBase1 {
	    initialization: string;
	    index_range: string;
	
	    static createFrom(source: any = {}) {
	        return new SegmentBase1(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.initialization = source["initialization"];
	        this.index_range = source["index_range"];
	    }
	}
	export class SegmentBase {
	    Initialization: string;
	    indexRange: string;
	
	    static createFrom(source: any = {}) {
	        return new SegmentBase(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Initialization = source["Initialization"];
	        this.indexRange = source["indexRange"];
	    }
	}
	export class AudioItem {
	    id: number;
	    baseUrl: string;
	    base_url: string;
	    backupUrl: string[];
	    backup_url: string[];
	    bandwidth: number;
	    mimeType: string;
	    mime_type: string;
	    codecs: string;
	    width: number;
	    height: number;
	    frameRate: string;
	    frame_rate: string;
	    sar: string;
	    startWithSap: number;
	    start_with_sap: number;
	    SegmentBase: SegmentBase;
	    segment_base: SegmentBase1;
	    codecid: number;
	
	    static createFrom(source: any = {}) {
	        return new AudioItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.baseUrl = source["baseUrl"];
	        this.base_url = source["base_url"];
	        this.backupUrl = source["backupUrl"];
	        this.backup_url = source["backup_url"];
	        this.bandwidth = source["bandwidth"];
	        this.mimeType = source["mimeType"];
	        this.mime_type = source["mime_type"];
	        this.codecs = source["codecs"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.frameRate = source["frameRate"];
	        this.frame_rate = source["frame_rate"];
	        this.sar = source["sar"];
	        this.startWithSap = source["startWithSap"];
	        this.start_with_sap = source["start_with_sap"];
	        this.SegmentBase = this.convertValues(source["SegmentBase"], SegmentBase);
	        this.segment_base = this.convertValues(source["segment_base"], SegmentBase1);
	        this.codecid = source["codecid"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AudioMetaUrlList {
	    backup_url: string;
	    fallback_url: string;
	    main_url: string;
	
	    static createFrom(source: any = {}) {
	        return new AudioMetaUrlList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.backup_url = source["backup_url"];
	        this.fallback_url = source["fallback_url"];
	        this.main_url = source["main_url"];
	    }
	}
	export class AudioMeta {
	    bitrate: number;
	    codec_type: string;
	    encoded_type: string;
	    file_hash: string;
	    file_cs: string;
	    format: string;
	    fps: number;
	    size: number;
	    media_type: string;
	    quality: string;
	    url_list: AudioMetaUrlList;
	
	    static createFrom(source: any = {}) {
	        return new AudioMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bitrate = source["bitrate"];
	        this.codec_type = source["codec_type"];
	        this.encoded_type = source["encoded_type"];
	        this.file_hash = source["file_hash"];
	        this.file_cs = source["file_cs"];
	        this.format = source["format"];
	        this.fps = source["fps"];
	        this.size = source["size"];
	        this.media_type = source["media_type"];
	        this.quality = source["quality"];
	        this.url_list = this.convertValues(source["url_list"], AudioMetaUrlList);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class AutoQnResp {
	    dyeid: string;
	
	    static createFrom(source: any = {}) {
	        return new AutoQnResp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dyeid = source["dyeid"];
	    }
	}
	export class Avatar {
	    height: number;
	    width: number;
	    uri: string;
	    url_list: string[];
	
	    static createFrom(source: any = {}) {
	        return new Avatar(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.height = source["height"];
	        this.width = source["width"];
	        this.uri = source["uri"];
	        this.url_list = source["url_list"];
	    }
	}
	export class AwemeAuthor {
	    cover_url: Avatar[];
	    avatar_thumb: Avatar;
	    follow_status: number;
	    follower_status: number;
	    nickname: string;
	    uid: string;
	    sec_uid: string;
	
	    static createFrom(source: any = {}) {
	        return new AwemeAuthor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cover_url = this.convertValues(source["cover_url"], Avatar);
	        this.avatar_thumb = this.convertValues(source["avatar_thumb"], Avatar);
	        this.follow_status = source["follow_status"];
	        this.follower_status = source["follower_status"];
	        this.nickname = source["nickname"];
	        this.uid = source["uid"];
	        this.sec_uid = source["sec_uid"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BitRateAudioItem {
	    audio_extra: string;
	    audio_meta: AudioMeta;
	    audio_quality: number;
	
	    static createFrom(source: any = {}) {
	        return new BitRateAudioItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.audio_extra = source["audio_extra"];
	        this.audio_meta = this.convertValues(source["audio_meta"], AudioMeta);
	        this.audio_quality = source["audio_quality"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BitRateItem {
	    FPS: number;
	    HDR_bit: string;
	    HDR_type: string;
	    bit_rate: number;
	    format: string;
	    gear_name: string;
	    is_bytevc1: number;
	    is_h265: number;
	    play_addr: PlayInfo;
	    quality_type: number;
	    video_extra: string;
	
	    static createFrom(source: any = {}) {
	        return new BitRateItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.FPS = source["FPS"];
	        this.HDR_bit = source["HDR_bit"];
	        this.HDR_type = source["HDR_type"];
	        this.bit_rate = source["bit_rate"];
	        this.format = source["format"];
	        this.gear_name = source["gear_name"];
	        this.is_bytevc1 = source["is_bytevc1"];
	        this.is_h265 = source["is_h265"];
	        this.play_addr = this.convertValues(source["play_addr"], PlayInfo);
	        this.quality_type = source["quality_type"];
	        this.video_extra = source["video_extra"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PlayInfo {
	    data_size: number;
	    file_cs: string;
	    file_hash: string;
	    height: number;
	    width: number;
	    uri: string;
	    url_key: string;
	    url_list: string[];
	
	    static createFrom(source: any = {}) {
	        return new PlayInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data_size = source["data_size"];
	        this.file_cs = source["file_cs"];
	        this.file_hash = source["file_hash"];
	        this.height = source["height"];
	        this.width = source["width"];
	        this.uri = source["uri"];
	        this.url_key = source["url_key"];
	        this.url_list = source["url_list"];
	    }
	}
	export class Cover {
	    uri: string;
	    url_list: string[];
	
	    static createFrom(source: any = {}) {
	        return new Cover(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uri = source["uri"];
	        this.url_list = source["url_list"];
	    }
	}
	export class Video {
	    cover: Cover;
	    duration: number;
	    format: string;
	    height: number;
	    width: number;
	    play_addr_h264: PlayInfo;
	    play_addr_265: PlayInfo;
	    play_addr: PlayInfo;
	    origin_cover: Cover;
	    gaussian_cover: Cover;
	    bit_rate: BitRateItem[];
	    bit_rate_audio: BitRateAudioItem[];
	
	    static createFrom(source: any = {}) {
	        return new Video(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cover = this.convertValues(source["cover"], Cover);
	        this.duration = source["duration"];
	        this.format = source["format"];
	        this.height = source["height"];
	        this.width = source["width"];
	        this.play_addr_h264 = this.convertValues(source["play_addr_h264"], PlayInfo);
	        this.play_addr_265 = this.convertValues(source["play_addr_265"], PlayInfo);
	        this.play_addr = this.convertValues(source["play_addr"], PlayInfo);
	        this.origin_cover = this.convertValues(source["origin_cover"], Cover);
	        this.gaussian_cover = this.convertValues(source["gaussian_cover"], Cover);
	        this.bit_rate = this.convertValues(source["bit_rate"], BitRateItem);
	        this.bit_rate_audio = this.convertValues(source["bit_rate_audio"], BitRateAudioItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class VideoStatistics {
	    collect_count: number;
	    comment_count: number;
	    digg_count: number;
	    recommend_count: number;
	    share_count: number;
	
	    static createFrom(source: any = {}) {
	        return new VideoStatistics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.collect_count = source["collect_count"];
	        this.comment_count = source["comment_count"];
	        this.digg_count = source["digg_count"];
	        this.recommend_count = source["recommend_count"];
	        this.share_count = source["share_count"];
	    }
	}
	export class AwemeItem {
	    authentication_token: string;
	    author: AwemeAuthor;
	    author_user_id: number;
	    caption: string;
	    desc: string;
	    item_title: string;
	    sec_item_id: string;
	    aweme_id: string;
	    group_id: string;
	    media_type: number;
	    create_time: number;
	    statistics: VideoStatistics;
	    video: Video;
	
	    static createFrom(source: any = {}) {
	        return new AwemeItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.authentication_token = source["authentication_token"];
	        this.author = this.convertValues(source["author"], AwemeAuthor);
	        this.author_user_id = source["author_user_id"];
	        this.caption = source["caption"];
	        this.desc = source["desc"];
	        this.item_title = source["item_title"];
	        this.sec_item_id = source["sec_item_id"];
	        this.aweme_id = source["aweme_id"];
	        this.group_id = source["group_id"];
	        this.media_type = source["media_type"];
	        this.create_time = source["create_time"];
	        this.statistics = this.convertValues(source["statistics"], VideoStatistics);
	        this.video = this.convertValues(source["video"], Video);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class UserInfo {
	    avatar_larger: Avatar;
	    avatar_medium: Avatar;
	    avatar_thumb: Avatar;
	    nickname: string;
	    uid: string;
	
	    static createFrom(source: any = {}) {
	        return new UserInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.avatar_larger = this.convertValues(source["avatar_larger"], Avatar);
	        this.avatar_medium = this.convertValues(source["avatar_medium"], Avatar);
	        this.avatar_thumb = this.convertValues(source["avatar_thumb"], Avatar);
	        this.nickname = source["nickname"];
	        this.uid = source["uid"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CollectsList {
	    app_id: number;
	    collects_cover: Cover;
	    collects_id: number;
	    collects_id_str: string;
	    collects_name: string;
	    create_time: number;
	    follow_status: number;
	    followed_count: number;
	    is_default_cover: boolean;
	    is_normal_status: boolean;
	    item_type: number;
	    last_collect_time: number;
	    play_count: number;
	    sicily_collects_cover_list: any;
	    states: number;
	    status: number;
	    system_type: number;
	    total_number: number;
	    user_id: number;
	    user_id_str: string;
	    user_info: UserInfo;
	
	    static createFrom(source: any = {}) {
	        return new CollectsList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.app_id = source["app_id"];
	        this.collects_cover = this.convertValues(source["collects_cover"], Cover);
	        this.collects_id = source["collects_id"];
	        this.collects_id_str = source["collects_id_str"];
	        this.collects_name = source["collects_name"];
	        this.create_time = source["create_time"];
	        this.follow_status = source["follow_status"];
	        this.followed_count = source["followed_count"];
	        this.is_default_cover = source["is_default_cover"];
	        this.is_normal_status = source["is_normal_status"];
	        this.item_type = source["item_type"];
	        this.last_collect_time = source["last_collect_time"];
	        this.play_count = source["play_count"];
	        this.sicily_collects_cover_list = source["sicily_collects_cover_list"];
	        this.states = source["states"];
	        this.status = source["status"];
	        this.system_type = source["system_type"];
	        this.total_number = source["total_number"];
	        this.user_id = source["user_id"];
	        this.user_id_str = source["user_id_str"];
	        this.user_info = this.convertValues(source["user_info"], UserInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CollectListResponse {
	    cursor: number;
	    total_number: number;
	    status_code: number;
	    has_more: boolean;
	    collects_list: CollectsList[];
	
	    static createFrom(source: any = {}) {
	        return new CollectListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cursor = source["cursor"];
	        this.total_number = source["total_number"];
	        this.status_code = source["status_code"];
	        this.has_more = source["has_more"];
	        this.collects_list = this.convertValues(source["collects_list"], CollectsList);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CollectionUpper {
	    mid: number;
	    name: string;
	    face: string;
	    jump_link: string;
	
	    static createFrom(source: any = {}) {
	        return new CollectionUpper(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mid = source["mid"];
	        this.name = source["name"];
	        this.face = source["face"];
	        this.jump_link = source["jump_link"];
	    }
	}
	export class CollectionDataList {
	    id: number;
	    fid: number;
	    mid: number;
	    attr: number;
	    attr_desc: string;
	    title: string;
	    cover: string;
	    upper: CollectionUpper;
	    cover_type: number;
	    intro: string;
	    ctime: number;
	    mtime: number;
	    state: number;
	    fav_state: number;
	    media_count: number;
	    view_count: number;
	    vt: number;
	    is_top: boolean;
	    recent_fav: any;
	    play_switch: number;
	    type: number;
	    link: string;
	    bvid: string;
	
	    static createFrom(source: any = {}) {
	        return new CollectionDataList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fid = source["fid"];
	        this.mid = source["mid"];
	        this.attr = source["attr"];
	        this.attr_desc = source["attr_desc"];
	        this.title = source["title"];
	        this.cover = source["cover"];
	        this.upper = this.convertValues(source["upper"], CollectionUpper);
	        this.cover_type = source["cover_type"];
	        this.intro = source["intro"];
	        this.ctime = source["ctime"];
	        this.mtime = source["mtime"];
	        this.state = source["state"];
	        this.fav_state = source["fav_state"];
	        this.media_count = source["media_count"];
	        this.view_count = source["view_count"];
	        this.vt = source["vt"];
	        this.is_top = source["is_top"];
	        this.recent_fav = source["recent_fav"];
	        this.play_switch = source["play_switch"];
	        this.type = source["type"];
	        this.link = source["link"];
	        this.bvid = source["bvid"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CollectionData {
	    count: number;
	    list: CollectionDataList[];
	    has_more: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CollectionData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.list = this.convertValues(source["list"], CollectionDataList);
	        this.has_more = source["has_more"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class CollectionItemCntInfo {
	    collect: number;
	    play: number;
	    danmaku: number;
	    vt: number;
	
	    static createFrom(source: any = {}) {
	        return new CollectionItemCntInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.collect = source["collect"];
	        this.play = source["play"];
	        this.danmaku = source["danmaku"];
	        this.vt = source["vt"];
	    }
	}
	export class CollectionItemMediaCntInfo {
	    collect: number;
	    play: number;
	    danmaku: number;
	    vt: number;
	
	    static createFrom(source: any = {}) {
	        return new CollectionItemMediaCntInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.collect = source["collect"];
	        this.play = source["play"];
	        this.danmaku = source["danmaku"];
	        this.vt = source["vt"];
	    }
	}
	export class CollectionItemMediaUpper {
	    mid: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new CollectionItemMediaUpper(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mid = source["mid"];
	        this.name = source["name"];
	    }
	}
	export class CollectionItemMedias {
	    id: number;
	    title: string;
	    cover: string;
	    duration: number;
	    pubtime: number;
	    bvid: string;
	    upper: CollectionItemMediaUpper;
	    cnt_info: CollectionItemMediaCntInfo;
	    enable_vt: number;
	    vt_display: string;
	    is_self_view: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CollectionItemMedias(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.cover = source["cover"];
	        this.duration = source["duration"];
	        this.pubtime = source["pubtime"];
	        this.bvid = source["bvid"];
	        this.upper = this.convertValues(source["upper"], CollectionItemMediaUpper);
	        this.cnt_info = this.convertValues(source["cnt_info"], CollectionItemMediaCntInfo);
	        this.enable_vt = source["enable_vt"];
	        this.vt_display = source["vt_display"];
	        this.is_self_view = source["is_self_view"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CollectionItemUpper {
	    mid: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new CollectionItemUpper(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mid = source["mid"];
	        this.name = source["name"];
	    }
	}
	export class CollectionItemInfo {
	    id: number;
	    season_type: number;
	    title: string;
	    cover: string;
	    upper: CollectionItemUpper;
	    cnt_info: CollectionItemCntInfo;
	    media_count: number;
	    intro: string;
	    enable_vt: number;
	
	    static createFrom(source: any = {}) {
	        return new CollectionItemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.season_type = source["season_type"];
	        this.title = source["title"];
	        this.cover = source["cover"];
	        this.upper = this.convertValues(source["upper"], CollectionItemUpper);
	        this.cnt_info = this.convertValues(source["cnt_info"], CollectionItemCntInfo);
	        this.media_count = source["media_count"];
	        this.intro = source["intro"];
	        this.enable_vt = source["enable_vt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CollectionItemData {
	    info: CollectionItemInfo;
	    medias: CollectionItemMedias[];
	
	    static createFrom(source: any = {}) {
	        return new CollectionItemData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.info = this.convertValues(source["info"], CollectionItemInfo);
	        this.medias = this.convertValues(source["medias"], CollectionItemMedias);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	
	
	
	
	export class Flac {
	    display: boolean;
	    audio: AudioItem;
	
	    static createFrom(source: any = {}) {
	        return new Flac(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.display = source["display"];
	        this.audio = this.convertValues(source["audio"], AudioItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Dolby {
	    type: number;
	    audio: any;
	
	    static createFrom(source: any = {}) {
	        return new Dolby(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.audio = source["audio"];
	    }
	}
	export class VideoItem {
	    id: number;
	    baseUrl: string;
	    base_url: string;
	    backupUrl: string[];
	    backup_url: string[];
	    bandwidth: number;
	    mimeType: string;
	    mime_type: string;
	    codecs: string;
	    width: number;
	    height: number;
	    frameRate: string;
	    frame_rate: string;
	    sar: string;
	    startWithSap: number;
	    start_with_sap: number;
	    SegmentBase: SegmentBase;
	    segment_base: SegmentBase1;
	    codecid: number;
	
	    static createFrom(source: any = {}) {
	        return new VideoItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.baseUrl = source["baseUrl"];
	        this.base_url = source["base_url"];
	        this.backupUrl = source["backupUrl"];
	        this.backup_url = source["backup_url"];
	        this.bandwidth = source["bandwidth"];
	        this.mimeType = source["mimeType"];
	        this.mime_type = source["mime_type"];
	        this.codecs = source["codecs"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.frameRate = source["frameRate"];
	        this.frame_rate = source["frame_rate"];
	        this.sar = source["sar"];
	        this.startWithSap = source["startWithSap"];
	        this.start_with_sap = source["start_with_sap"];
	        this.SegmentBase = this.convertValues(source["SegmentBase"], SegmentBase);
	        this.segment_base = this.convertValues(source["segment_base"], SegmentBase1);
	        this.codecid = source["codecid"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Dash {
	    duration: number;
	    minBufferTime: number;
	    min_buffer_time: number;
	    video: VideoItem[];
	    audio: AudioItem[];
	    dolby: Dolby;
	    flac?: Flac;
	
	    static createFrom(source: any = {}) {
	        return new Dash(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.duration = source["duration"];
	        this.minBufferTime = source["minBufferTime"];
	        this.min_buffer_time = source["min_buffer_time"];
	        this.video = this.convertValues(source["video"], VideoItem);
	        this.audio = this.convertValues(source["audio"], AudioItem);
	        this.dolby = this.convertValues(source["dolby"], Dolby);
	        this.flac = this.convertValues(source["flac"], Flac);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class UgcSeasonPageItem {
	    cid: number;
	    page: number;
	    from: string;
	    part: string;
	    duration: number;
	    vid: string;
	    weblink: string;
	    dimension: Dimension;
	
	    static createFrom(source: any = {}) {
	        return new UgcSeasonPageItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cid = source["cid"];
	        this.page = source["page"];
	        this.from = source["from"];
	        this.part = source["part"];
	        this.duration = source["duration"];
	        this.vid = source["vid"];
	        this.weblink = source["weblink"];
	        this.dimension = this.convertValues(source["dimension"], Dimension);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Episode {
	    season_id: number;
	    section_id: number;
	    id: number;
	    aid: number;
	    cid: number;
	    title: string;
	    attribute: number;
	    arc: Arc;
	    page: UgcSeasonPageItem;
	    bvid: string;
	    pages: UgcSeasonPageItem[];
	
	    static createFrom(source: any = {}) {
	        return new Episode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.season_id = source["season_id"];
	        this.section_id = source["section_id"];
	        this.id = source["id"];
	        this.aid = source["aid"];
	        this.cid = source["cid"];
	        this.title = source["title"];
	        this.attribute = source["attribute"];
	        this.arc = this.convertValues(source["arc"], Arc);
	        this.page = this.convertValues(source["page"], UgcSeasonPageItem);
	        this.bvid = source["bvid"];
	        this.pages = this.convertValues(source["pages"], UgcSeasonPageItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EpisodicButton {
	    text: string;
	    uri: string;
	
	    static createFrom(source: any = {}) {
	        return new EpisodicButton(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.uri = source["uri"];
	    }
	}
	export class FavoriteMediaUgc {
	    first_cid: number;
	
	    static createFrom(source: any = {}) {
	        return new FavoriteMediaUgc(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.first_cid = source["first_cid"];
	    }
	}
	export class FavoriteMediaCntInfo {
	    collect: number;
	    play: number;
	    danmaku: number;
	    vt: number;
	    play_switch: number;
	    reply: number;
	    view_text_1: string;
	
	    static createFrom(source: any = {}) {
	        return new FavoriteMediaCntInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.collect = source["collect"];
	        this.play = source["play"];
	        this.danmaku = source["danmaku"];
	        this.vt = source["vt"];
	        this.play_switch = source["play_switch"];
	        this.reply = source["reply"];
	        this.view_text_1 = source["view_text_1"];
	    }
	}
	export class FavoriteMediaUpper {
	    mid: number;
	    name: string;
	    face: string;
	    jump_link: string;
	
	    static createFrom(source: any = {}) {
	        return new FavoriteMediaUpper(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mid = source["mid"];
	        this.name = source["name"];
	        this.face = source["face"];
	        this.jump_link = source["jump_link"];
	    }
	}
	export class FavoriteMedias {
	    id: number;
	    type: number;
	    title: string;
	    cover: string;
	    intro: string;
	    page: number;
	    duration: number;
	    upper: FavoriteMediaUpper;
	    attr: number;
	    cnt_info: FavoriteMediaCntInfo;
	    link: string;
	    ctime: number;
	    pubtime: number;
	    fav_time: number;
	    bv_id: string;
	    bvid: string;
	    season: any;
	    ogv: any;
	    ugc: FavoriteMediaUgc;
	    media_list_link: string;
	
	    static createFrom(source: any = {}) {
	        return new FavoriteMedias(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.title = source["title"];
	        this.cover = source["cover"];
	        this.intro = source["intro"];
	        this.page = source["page"];
	        this.duration = source["duration"];
	        this.upper = this.convertValues(source["upper"], FavoriteMediaUpper);
	        this.attr = source["attr"];
	        this.cnt_info = this.convertValues(source["cnt_info"], FavoriteMediaCntInfo);
	        this.link = source["link"];
	        this.ctime = source["ctime"];
	        this.pubtime = source["pubtime"];
	        this.fav_time = source["fav_time"];
	        this.bv_id = source["bv_id"];
	        this.bvid = source["bvid"];
	        this.season = source["season"];
	        this.ogv = source["ogv"];
	        this.ugc = this.convertValues(source["ugc"], FavoriteMediaUgc);
	        this.media_list_link = source["media_list_link"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FavoriteInfoCntInfo {
	    collect: number;
	    play: number;
	    thumb_up: number;
	    share: number;
	
	    static createFrom(source: any = {}) {
	        return new FavoriteInfoCntInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.collect = source["collect"];
	        this.play = source["play"];
	        this.thumb_up = source["thumb_up"];
	        this.share = source["share"];
	    }
	}
	export class FavoriteInfoUpper {
	    mid: number;
	    name: string;
	    face: string;
	    followed: boolean;
	    vip_type: number;
	    vip_statue: number;
	
	    static createFrom(source: any = {}) {
	        return new FavoriteInfoUpper(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mid = source["mid"];
	        this.name = source["name"];
	        this.face = source["face"];
	        this.followed = source["followed"];
	        this.vip_type = source["vip_type"];
	        this.vip_statue = source["vip_statue"];
	    }
	}
	export class FavoriteInfo {
	    id: number;
	    fid: number;
	    mid: number;
	    attr: number;
	    title: string;
	    cover: string;
	    upper: FavoriteInfoUpper;
	    cover_type: number;
	    cnt_info: FavoriteInfoCntInfo;
	    type: number;
	    intro: string;
	    ctime: number;
	    mtime: number;
	    state: number;
	    fav_state: number;
	    like_state: number;
	    media_count: number;
	    is_top: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FavoriteInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fid = source["fid"];
	        this.mid = source["mid"];
	        this.attr = source["attr"];
	        this.title = source["title"];
	        this.cover = source["cover"];
	        this.upper = this.convertValues(source["upper"], FavoriteInfoUpper);
	        this.cover_type = source["cover_type"];
	        this.cnt_info = this.convertValues(source["cnt_info"], FavoriteInfoCntInfo);
	        this.type = source["type"];
	        this.intro = source["intro"];
	        this.ctime = source["ctime"];
	        this.mtime = source["mtime"];
	        this.state = source["state"];
	        this.fav_state = source["fav_state"];
	        this.like_state = source["like_state"];
	        this.media_count = source["media_count"];
	        this.is_top = source["is_top"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FavoriteData {
	    info: FavoriteInfo;
	    medias: FavoriteMedias[];
	    has_more: boolean;
	    ttl: number;
	
	    static createFrom(source: any = {}) {
	        return new FavoriteData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.info = this.convertValues(source["info"], FavoriteInfo);
	        this.medias = this.convertValues(source["medias"], FavoriteMedias);
	        this.has_more = source["has_more"];
	        this.ttl = source["ttl"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class FavoriteItem {
	    id: number;
	    fid: number;
	    mid: number;
	    attr: number;
	    title: string;
	    fav_state: number;
	    media_count: number;
	
	    static createFrom(source: any = {}) {
	        return new FavoriteItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fid = source["fid"];
	        this.mid = source["mid"];
	        this.attr = source["attr"];
	        this.title = source["title"];
	        this.fav_state = source["fav_state"];
	        this.media_count = source["media_count"];
	    }
	}
	
	
	
	
	export class FavoriteVideoResponse {
	    cursor: number;
	    has_more: number;
	    status_code: number;
	    uid: string;
	    sec_uid: string;
	    aweme_list: AwemeItem[];
	
	    static createFrom(source: any = {}) {
	        return new FavoriteVideoResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cursor = source["cursor"];
	        this.has_more = source["has_more"];
	        this.status_code = source["status_code"];
	        this.uid = source["uid"];
	        this.sec_uid = source["sec_uid"];
	        this.aweme_list = this.convertValues(source["aweme_list"], AwemeItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FavoritesData {
	    count: number;
	    list: FavoriteItem[];
	    season: any;
	
	    static createFrom(source: any = {}) {
	        return new FavoritesData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.list = this.convertValues(source["list"], FavoriteItem);
	        this.season = source["season"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class FollowDataList {
	    mid: number;
	    attribute: number;
	    mtime: number;
	    tag: any;
	    special: number;
	    uname: string;
	    face: string;
	    sign: string;
	
	    static createFrom(source: any = {}) {
	        return new FollowDataList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mid = source["mid"];
	        this.attribute = source["attribute"];
	        this.mtime = source["mtime"];
	        this.tag = source["tag"];
	        this.special = source["special"];
	        this.uname = source["uname"];
	        this.face = source["face"];
	        this.sign = source["sign"];
	    }
	}
	export class FollowData {
	    list: FollowDataList[];
	    re_version: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new FollowData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.list = this.convertValues(source["list"], FollowDataList);
	        this.re_version = source["re_version"];
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class FollowItem {
	    signature: string;
	    aweme_count: number;
	    nickname: string;
	    sec_uid: string;
	    uid: string;
	    unique_id: string;
	    short_id: string;
	    avatar_larger: Avatar;
	    avatar_medium: Avatar;
	    avatar_thumb: Avatar;
	    follower_count: number;
	
	    static createFrom(source: any = {}) {
	        return new FollowItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.signature = source["signature"];
	        this.aweme_count = source["aweme_count"];
	        this.nickname = source["nickname"];
	        this.sec_uid = source["sec_uid"];
	        this.uid = source["uid"];
	        this.unique_id = source["unique_id"];
	        this.short_id = source["short_id"];
	        this.avatar_larger = this.convertValues(source["avatar_larger"], Avatar);
	        this.avatar_medium = this.convertValues(source["avatar_medium"], Avatar);
	        this.avatar_thumb = this.convertValues(source["avatar_thumb"], Avatar);
	        this.follower_count = source["follower_count"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FollowResponse {
	    followings: FollowItem[];
	    has_more: boolean;
	    status_code: number;
	    total: number;
	    mix_count: number;
	    myself_user_id: string;
	    offset: number;
	
	    static createFrom(source: any = {}) {
	        return new FollowResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.followings = this.convertValues(source["followings"], FollowItem);
	        this.has_more = source["has_more"];
	        this.status_code = source["status_code"];
	        this.total = source["total"];
	        this.mix_count = source["mix_count"];
	        this.myself_user_id = source["myself_user_id"];
	        this.offset = source["offset"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LogOutData {
	    redirectUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new LogOutData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.redirectUrl = source["redirectUrl"];
	    }
	}
	export class LogOut {
	    code: number;
	    status: boolean;
	    ts: number;
	    data: LogOutData;
	
	    static createFrom(source: any = {}) {
	        return new LogOut(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.status = source["status"];
	        this.ts = source["ts"];
	        this.data = this.convertValues(source["data"], LogOutData);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class MyInfoData {
	    nickname: string;
	    sec_uid: string;
	    short_id: string;
	    signature: string;
	    uid: string;
	    unique_id: string;
	    avatar_thumb: Avatar;
	
	    static createFrom(source: any = {}) {
	        return new MyInfoData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nickname = source["nickname"];
	        this.sec_uid = source["sec_uid"];
	        this.short_id = source["short_id"];
	        this.signature = source["signature"];
	        this.uid = source["uid"];
	        this.unique_id = source["unique_id"];
	        this.avatar_thumb = this.convertValues(source["avatar_thumb"], Avatar);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MyInfo {
	    owner_sec_uid: string;
	    status_code: number;
	    next_req_count: number;
	    followings: MyInfoData[];
	
	    static createFrom(source: any = {}) {
	        return new MyInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.owner_sec_uid = source["owner_sec_uid"];
	        this.status_code = source["status_code"];
	        this.next_req_count = source["next_req_count"];
	        this.followings = this.convertValues(source["followings"], MyInfoData);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class MyInfoExpertInfo {
	    title: string;
	    state: number;
	    type: number;
	    desc: string;
	
	    static createFrom(source: any = {}) {
	        return new MyInfoExpertInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.state = source["state"];
	        this.type = source["type"];
	        this.desc = source["desc"];
	    }
	}
	export class MyInfoNamePlate {
	    nid: number;
	    name: string;
	    image: string;
	    image_small: string;
	    level: string;
	    condition: string;
	
	    static createFrom(source: any = {}) {
	        return new MyInfoNamePlate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nid = source["nid"];
	        this.name = source["name"];
	        this.image = source["image"];
	        this.image_small = source["image_small"];
	        this.level = source["level"];
	        this.condition = source["condition"];
	    }
	}
	export class MyInfoOfficial {
	    role: number;
	    title: string;
	    desc: string;
	    type: number;
	
	    static createFrom(source: any = {}) {
	        return new MyInfoOfficial(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.title = source["title"];
	        this.desc = source["desc"];
	        this.type = source["type"];
	    }
	}
	export class MyInfoProfession {
	    id: number;
	    name: string;
	    is_show: number;
	    category_one: string;
	    realname: string;
	    title: string;
	    department: string;
	    certificate_no: string;
	    certificate_show: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MyInfoProfession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.is_show = source["is_show"];
	        this.category_one = source["category_one"];
	        this.realname = source["realname"];
	        this.title = source["title"];
	        this.department = source["department"];
	        this.certificate_no = source["certificate_no"];
	        this.certificate_show = source["certificate_show"];
	    }
	}
	export class MyInfoVipSuperVip {
	    is_super_vip: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MyInfoVipSuperVip(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_super_vip = source["is_super_vip"];
	    }
	}
	export class MyInfoVipOttInfo {
	    vip_type: number;
	    pay_type: number;
	    pay_channel_id: string;
	    status: number;
	    overdue_time: number;
	
	    static createFrom(source: any = {}) {
	        return new MyInfoVipOttInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vip_type = source["vip_type"];
	        this.pay_type = source["pay_type"];
	        this.pay_channel_id = source["pay_channel_id"];
	        this.status = source["status"];
	        this.overdue_time = source["overdue_time"];
	    }
	}
	export class MyInfoVipLabel {
	    path: string;
	    text: string;
	    label_theme: string;
	    text_color: string;
	    bg_color: string;
	    border_color: string;
	    img_label_uri_hant_static: string;
	    img_label_uri_hans_static: string;
	    label_id: number;
	    bg_style: number;
	    use_img_label: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MyInfoVipLabel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.text = source["text"];
	        this.label_theme = source["label_theme"];
	        this.text_color = source["text_color"];
	        this.bg_color = source["bg_color"];
	        this.border_color = source["border_color"];
	        this.img_label_uri_hant_static = source["img_label_uri_hant_static"];
	        this.img_label_uri_hans_static = source["img_label_uri_hans_static"];
	        this.label_id = source["label_id"];
	        this.bg_style = source["bg_style"];
	        this.use_img_label = source["use_img_label"];
	    }
	}
	export class UserVip {
	    type: number;
	    status: number;
	    due_date: number;
	    theme_type: number;
	    label: MyInfoVipLabel;
	    avatar_subscript: number;
	    nickname_color: string;
	    OttInfo: MyInfoVipOttInfo;
	    super_vip: MyInfoVipSuperVip;
	    role: number;
	
	    static createFrom(source: any = {}) {
	        return new UserVip(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.status = source["status"];
	        this.due_date = source["due_date"];
	        this.theme_type = source["theme_type"];
	        this.label = this.convertValues(source["label"], MyInfoVipLabel);
	        this.avatar_subscript = source["avatar_subscript"];
	        this.nickname_color = source["nickname_color"];
	        this.OttInfo = this.convertValues(source["OttInfo"], MyInfoVipOttInfo);
	        this.super_vip = this.convertValues(source["super_vip"], MyInfoVipSuperVip);
	        this.role = source["role"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MyInfoProfile {
	    mid: number;
	    name: string;
	    sex: string;
	    face: string;
	    sign: string;
	    rank: number;
	    level: number;
	    birthday: number;
	    jointime: number;
	    moral: number;
	    silence: number;
	    email_status: number;
	    tel_status: number;
	    identification: number;
	    is_fake_account: number;
	    is_tourist: number;
	    pin_prompting: number;
	    official: MyInfoOfficial;
	    nameplate: MyInfoNamePlate;
	    vip: UserVip;
	    is_rip_user: boolean;
	    is_reg_audit: number;
	    country_code: string;
	    expert_info: MyInfoExpertInfo;
	    profession: MyInfoProfession;
	
	    static createFrom(source: any = {}) {
	        return new MyInfoProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mid = source["mid"];
	        this.name = source["name"];
	        this.sex = source["sex"];
	        this.face = source["face"];
	        this.sign = source["sign"];
	        this.rank = source["rank"];
	        this.level = source["level"];
	        this.birthday = source["birthday"];
	        this.jointime = source["jointime"];
	        this.moral = source["moral"];
	        this.silence = source["silence"];
	        this.email_status = source["email_status"];
	        this.tel_status = source["tel_status"];
	        this.identification = source["identification"];
	        this.is_fake_account = source["is_fake_account"];
	        this.is_tourist = source["is_tourist"];
	        this.pin_prompting = source["pin_prompting"];
	        this.official = this.convertValues(source["official"], MyInfoOfficial);
	        this.nameplate = this.convertValues(source["nameplate"], MyInfoNamePlate);
	        this.vip = this.convertValues(source["vip"], UserVip);
	        this.is_rip_user = source["is_rip_user"];
	        this.is_reg_audit = source["is_reg_audit"];
	        this.country_code = source["country_code"];
	        this.expert_info = this.convertValues(source["expert_info"], MyInfoExpertInfo);
	        this.profession = this.convertValues(source["profession"], MyInfoProfession);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class PlayConf {
	    is_new_description: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PlayConf(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_new_description = source["is_new_description"];
	    }
	}
	
	export class PollQRCodeData {
	    url: string;
	    refresh_token: string;
	    timestamp: number;
	    code: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new PollQRCodeData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.refresh_token = source["refresh_token"];
	        this.timestamp = source["timestamp"];
	        this.code = source["code"];
	        this.message = source["message"];
	    }
	}
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
	export class RefreshData {
	    refresh: boolean;
	    timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new RefreshData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.refresh = source["refresh"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class SeasonsSeriesArchivesStat {
	    view: number;
	    vt: number;
	    danmaku: number;
	
	    static createFrom(source: any = {}) {
	        return new SeasonsSeriesArchivesStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.view = source["view"];
	        this.vt = source["vt"];
	        this.danmaku = source["danmaku"];
	    }
	}
	export class SeasonsArchivesItem {
	    aid: number;
	    bvid: string;
	    ctime: number;
	    duration: number;
	    enable_vt: boolean;
	    interactive_video: boolean;
	    pic: string;
	    playback_position: number;
	    pubdate: number;
	    state: number;
	    stat: SeasonsSeriesArchivesStat;
	    title: string;
	    ugc_pay: number;
	    vt_display: string;
	    is_lesson_video: number;
	
	    static createFrom(source: any = {}) {
	        return new SeasonsArchivesItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.aid = source["aid"];
	        this.bvid = source["bvid"];
	        this.ctime = source["ctime"];
	        this.duration = source["duration"];
	        this.enable_vt = source["enable_vt"];
	        this.interactive_video = source["interactive_video"];
	        this.pic = source["pic"];
	        this.playback_position = source["playback_position"];
	        this.pubdate = source["pubdate"];
	        this.state = source["state"];
	        this.stat = this.convertValues(source["stat"], SeasonsSeriesArchivesStat);
	        this.title = source["title"];
	        this.ugc_pay = source["ugc_pay"];
	        this.vt_display = source["vt_display"];
	        this.is_lesson_video = source["is_lesson_video"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SeasonsArchivesMeta {
	    cover: string;
	    description: string;
	    title: string;
	    category: number;
	    mid: number;
	    ptime: number;
	    season_id: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new SeasonsArchivesMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cover = source["cover"];
	        this.description = source["description"];
	        this.title = source["title"];
	        this.category = source["category"];
	        this.mid = source["mid"];
	        this.ptime = source["ptime"];
	        this.season_id = source["season_id"];
	        this.total = source["total"];
	    }
	}
	export class SeasonsArchivesPage {
	    PageNum: number;
	    PageSize: number;
	    Total: number;
	
	    static createFrom(source: any = {}) {
	        return new SeasonsArchivesPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PageNum = source["PageNum"];
	        this.PageSize = source["PageSize"];
	        this.Total = source["Total"];
	    }
	}
	export class SeasonsArchivesData {
	    aids: number[];
	    page: SeasonsArchivesPage;
	    meta: SeasonsArchivesMeta;
	    archives: SeasonsArchivesItem[];
	
	    static createFrom(source: any = {}) {
	        return new SeasonsArchivesData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.aids = source["aids"];
	        this.page = this.convertValues(source["page"], SeasonsArchivesPage);
	        this.meta = this.convertValues(source["meta"], SeasonsArchivesMeta);
	        this.archives = this.convertValues(source["archives"], SeasonsArchivesItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class SeasonsItemMeta {
	    category: number;
	    cover: string;
	    description: string;
	    mid: number;
	    name: string;
	    ptime: number;
	    season_id: number;
	    total: number;
	    title: string;
	
	    static createFrom(source: any = {}) {
	        return new SeasonsItemMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.cover = source["cover"];
	        this.description = source["description"];
	        this.mid = source["mid"];
	        this.name = source["name"];
	        this.ptime = source["ptime"];
	        this.season_id = source["season_id"];
	        this.total = source["total"];
	        this.title = source["title"];
	    }
	}
	export class SeasonsSeriesArchivesItem {
	    aid: number;
	    bvid: string;
	    ctime: number;
	    duration: number;
	    enable_vt: boolean;
	    interactive_video: boolean;
	    pic: string;
	    playback_position: number;
	    pubdate: number;
	    stat: SeasonsSeriesArchivesStat;
	    state: number;
	    title: string;
	    ugc_pay: number;
	    vt_display: string;
	    is_lesson_video: number;
	
	    static createFrom(source: any = {}) {
	        return new SeasonsSeriesArchivesItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.aid = source["aid"];
	        this.bvid = source["bvid"];
	        this.ctime = source["ctime"];
	        this.duration = source["duration"];
	        this.enable_vt = source["enable_vt"];
	        this.interactive_video = source["interactive_video"];
	        this.pic = source["pic"];
	        this.playback_position = source["playback_position"];
	        this.pubdate = source["pubdate"];
	        this.stat = this.convertValues(source["stat"], SeasonsSeriesArchivesStat);
	        this.state = source["state"];
	        this.title = source["title"];
	        this.ugc_pay = source["ugc_pay"];
	        this.vt_display = source["vt_display"];
	        this.is_lesson_video = source["is_lesson_video"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SeasonsItem {
	    archives: SeasonsSeriesArchivesItem[];
	    meta: SeasonsItemMeta;
	    recent_aids: number[];
	
	    static createFrom(source: any = {}) {
	        return new SeasonsItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.archives = this.convertValues(source["archives"], SeasonsSeriesArchivesItem);
	        this.meta = this.convertValues(source["meta"], SeasonsItemMeta);
	        this.recent_aids = source["recent_aids"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class SeriesItemMeta {
	    category: number;
	    cover: string;
	    creator: string;
	    ctime: number;
	    description: string;
	    keywords: string[];
	    last_update_ts: number;
	    mid: number;
	    mtime: number;
	    name: string;
	    raw_keywords: string;
	    series_id: number;
	    state: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new SeriesItemMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.cover = source["cover"];
	        this.creator = source["creator"];
	        this.ctime = source["ctime"];
	        this.description = source["description"];
	        this.keywords = source["keywords"];
	        this.last_update_ts = source["last_update_ts"];
	        this.mid = source["mid"];
	        this.mtime = source["mtime"];
	        this.name = source["name"];
	        this.raw_keywords = source["raw_keywords"];
	        this.series_id = source["series_id"];
	        this.state = source["state"];
	        this.total = source["total"];
	    }
	}
	export class SeriesItem {
	    archives: SeasonsSeriesArchivesItem[];
	    meta: SeriesItemMeta;
	    recent_aids: number[];
	
	    static createFrom(source: any = {}) {
	        return new SeriesItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.archives = this.convertValues(source["archives"], SeasonsSeriesArchivesItem);
	        this.meta = this.convertValues(source["meta"], SeriesItemMeta);
	        this.recent_aids = source["recent_aids"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SeasonsSeriesPage {
	    page_num: number;
	    page_size: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new SeasonsSeriesPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page_num = source["page_num"];
	        this.page_size = source["page_size"];
	        this.total = source["total"];
	    }
	}
	export class SeasonsSeriesItemsLists {
	    page: SeasonsSeriesPage;
	    seasons_list: SeasonsItem[];
	    series_list: SeriesItem[];
	
	    static createFrom(source: any = {}) {
	        return new SeasonsSeriesItemsLists(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = this.convertValues(source["page"], SeasonsSeriesPage);
	        this.seasons_list = this.convertValues(source["seasons_list"], SeasonsItem);
	        this.series_list = this.convertValues(source["series_list"], SeriesItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SeasonsSeriesData {
	    items_lists: SeasonsSeriesItemsLists;
	
	    static createFrom(source: any = {}) {
	        return new SeasonsSeriesData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items_lists = this.convertValues(source["items_lists"], SeasonsSeriesItemsLists);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	export class SeriesArchivesItem {
	    aid: number;
	    title: string;
	    pubdate: number;
	    ctime: number;
	    state: number;
	    pic: string;
	    duration: number;
	    bvid: string;
	    ugc_pay: number;
	    interactive_video: boolean;
	    enable_vt: number;
	    vt_display: string;
	    desc: string;
	    up_mid: number;
	    stat: SeasonsSeriesArchivesStat;
	
	    static createFrom(source: any = {}) {
	        return new SeriesArchivesItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.aid = source["aid"];
	        this.title = source["title"];
	        this.pubdate = source["pubdate"];
	        this.ctime = source["ctime"];
	        this.state = source["state"];
	        this.pic = source["pic"];
	        this.duration = source["duration"];
	        this.bvid = source["bvid"];
	        this.ugc_pay = source["ugc_pay"];
	        this.interactive_video = source["interactive_video"];
	        this.enable_vt = source["enable_vt"];
	        this.vt_display = source["vt_display"];
	        this.desc = source["desc"];
	        this.up_mid = source["up_mid"];
	        this.stat = this.convertValues(source["stat"], SeasonsSeriesArchivesStat);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SeriesListPage {
	    num: number;
	    size: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new SeriesListPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.num = source["num"];
	        this.size = source["size"];
	        this.total = source["total"];
	    }
	}
	export class SeriesArchivesData {
	    aids: number[];
	    page: SeriesListPage;
	    archives: SeriesArchivesItem[];
	
	    static createFrom(source: any = {}) {
	        return new SeriesArchivesData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.aids = source["aids"];
	        this.page = this.convertValues(source["page"], SeriesListPage);
	        this.archives = this.convertValues(source["archives"], SeriesArchivesItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	export class SupportFormat {
	    quality: number;
	    format: string;
	    new_description: string;
	    display_desc: string;
	    superscript: string;
	    codecs: string[];
	    can_watch_qn_reason: number;
	    limit_watch_reason: number;
	    report: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new SupportFormat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.quality = source["quality"];
	        this.format = source["format"];
	        this.new_description = source["new_description"];
	        this.display_desc = source["display_desc"];
	        this.superscript = source["superscript"];
	        this.codecs = source["codecs"];
	        this.can_watch_qn_reason = source["can_watch_qn_reason"];
	        this.limit_watch_reason = source["limit_watch_reason"];
	        this.report = source["report"];
	    }
	}
	export class UgcSeasonStat {
	    season_id: number;
	    view: number;
	    danmaku: number;
	    reply: number;
	    fav: number;
	    coin: number;
	    share: number;
	    now_rank: number;
	    his_rank: number;
	    like: number;
	    vt: number;
	    vv: number;
	
	    static createFrom(source: any = {}) {
	        return new UgcSeasonStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.season_id = source["season_id"];
	        this.view = source["view"];
	        this.danmaku = source["danmaku"];
	        this.reply = source["reply"];
	        this.fav = source["fav"];
	        this.coin = source["coin"];
	        this.share = source["share"];
	        this.now_rank = source["now_rank"];
	        this.his_rank = source["his_rank"];
	        this.like = source["like"];
	        this.vt = source["vt"];
	        this.vv = source["vv"];
	    }
	}
	export class UgcSeasonSectionItem {
	    season_id: number;
	    id: number;
	    title: string;
	    type: number;
	    episodes: Episode[];
	
	    static createFrom(source: any = {}) {
	        return new UgcSeasonSectionItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.season_id = source["season_id"];
	        this.id = source["id"];
	        this.title = source["title"];
	        this.type = source["type"];
	        this.episodes = this.convertValues(source["episodes"], Episode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UgcSeason {
	    id: number;
	    title: string;
	    cover: string;
	    mid: number;
	    intro: string;
	    sign_state: number;
	    attribute: number;
	    sections: UgcSeasonSectionItem[];
	    stat: UgcSeasonStat;
	    ep_count: number;
	    season_type: number;
	    is_pay_season: boolean;
	    enable_vt: number;
	
	    static createFrom(source: any = {}) {
	        return new UgcSeason(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.cover = source["cover"];
	        this.mid = source["mid"];
	        this.intro = source["intro"];
	        this.sign_state = source["sign_state"];
	        this.attribute = source["attribute"];
	        this.sections = this.convertValues(source["sections"], UgcSeasonSectionItem);
	        this.stat = this.convertValues(source["stat"], UgcSeasonStat);
	        this.ep_count = source["ep_count"];
	        this.season_type = source["season_type"];
	        this.is_pay_season = source["is_pay_season"];
	        this.enable_vt = source["enable_vt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class User {
	    avatar_larger: Avatar;
	    avatar_medium: Avatar;
	    avatar_thumb: Avatar;
	    city: string;
	    aweme_count: number;
	    country: string;
	    follower_count: number;
	    ip_location: string;
	    nickname: string;
	    province: string;
	    signature: string;
	    uid: string;
	    unique_id: string;
	    sec_uid: string;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.avatar_larger = this.convertValues(source["avatar_larger"], Avatar);
	        this.avatar_medium = this.convertValues(source["avatar_medium"], Avatar);
	        this.avatar_thumb = this.convertValues(source["avatar_thumb"], Avatar);
	        this.city = source["city"];
	        this.aweme_count = source["aweme_count"];
	        this.country = source["country"];
	        this.follower_count = source["follower_count"];
	        this.ip_location = source["ip_location"];
	        this.nickname = source["nickname"];
	        this.province = source["province"];
	        this.signature = source["signature"];
	        this.uid = source["uid"];
	        this.unique_id = source["unique_id"];
	        this.sec_uid = source["sec_uid"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UserData {
	    user: User;
	    status_code: number;
	
	    static createFrom(source: any = {}) {
	        return new UserData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user = this.convertValues(source["user"], User);
	        this.status_code = source["status_code"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UserGarb {
	    url_image_ani_cut: string;
	
	    static createFrom(source: any = {}) {
	        return new UserGarb(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url_image_ani_cut = source["url_image_ani_cut"];
	    }
	}
	
	export class UserInfoData {
	    mid: number;
	    name: string;
	    sex: string;
	    face: string;
	    sign: string;
	    rank: number;
	    level: number;
	    birthday: string;
	    is_followed: boolean;
	    is_risk: boolean;
	    vip: UserVip;
	
	    static createFrom(source: any = {}) {
	        return new UserInfoData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mid = source["mid"];
	        this.name = source["name"];
	        this.sex = source["sex"];
	        this.face = source["face"];
	        this.sign = source["sign"];
	        this.rank = source["rank"];
	        this.level = source["level"];
	        this.birthday = source["birthday"];
	        this.is_followed = source["is_followed"];
	        this.is_risk = source["is_risk"];
	        this.vip = this.convertValues(source["vip"], UserVip);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class VListItemMetaStat {
	    season_id: number;
	    view: number;
	    danmaku: number;
	    reply: number;
	    favorite: number;
	    coin: number;
	    share: number;
	    like: number;
	    mtime: number;
	    vt: number;
	    vv: number;
	
	    static createFrom(source: any = {}) {
	        return new VListItemMetaStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.season_id = source["season_id"];
	        this.view = source["view"];
	        this.danmaku = source["danmaku"];
	        this.reply = source["reply"];
	        this.favorite = source["favorite"];
	        this.coin = source["coin"];
	        this.share = source["share"];
	        this.like = source["like"];
	        this.mtime = source["mtime"];
	        this.vt = source["vt"];
	        this.vv = source["vv"];
	    }
	}
	export class VideoVListItemMeta {
	    id: number;
	    title: string;
	    cover: string;
	    mid: number;
	    intro: string;
	    sign_state: number;
	    attribute: number;
	    stat: VListItemMetaStat;
	    ep_count: number;
	    first_aid: number;
	    ptime: number;
	    ep_num: number;
	    show: number;
	
	    static createFrom(source: any = {}) {
	        return new VideoVListItemMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.cover = source["cover"];
	        this.mid = source["mid"];
	        this.intro = source["intro"];
	        this.sign_state = source["sign_state"];
	        this.attribute = source["attribute"];
	        this.stat = this.convertValues(source["stat"], VListItemMetaStat);
	        this.ep_count = source["ep_count"];
	        this.first_aid = source["first_aid"];
	        this.ptime = source["ptime"];
	        this.ep_num = source["ep_num"];
	        this.show = source["show"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class VListItem {
	    comment: number;
	    typeid: number;
	    play: number;
	    pic: string;
	    subtitle: string;
	    description: string;
	    copyright: string;
	    title: string;
	    review: number;
	    author: string;
	    mid: number;
	    created: number;
	    length: string;
	    video_review: number;
	    aid: number;
	    bvid: string;
	    hide_click: boolean;
	    is_pay: number;
	    is_union_video: number;
	    is_steins_gate: number;
	    is_live_playback: number;
	    is_lesson_video: number;
	    is_lesson_finished: number;
	    lesson_update_info: string;
	    jump_url: string;
	    meta: VideoVListItemMeta;
	    is_avoided: number;
	    season_id: number;
	    attribute: number;
	    is_charging_arc: boolean;
	    elec_arc_type: number;
	    elec_arc_badge: string;
	    vt: number;
	    enable_vt: number;
	    vt_display: string;
	    playback_position: number;
	    is_self_view: boolean;
	    view_self_type: number;
	
	    static createFrom(source: any = {}) {
	        return new VListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.comment = source["comment"];
	        this.typeid = source["typeid"];
	        this.play = source["play"];
	        this.pic = source["pic"];
	        this.subtitle = source["subtitle"];
	        this.description = source["description"];
	        this.copyright = source["copyright"];
	        this.title = source["title"];
	        this.review = source["review"];
	        this.author = source["author"];
	        this.mid = source["mid"];
	        this.created = source["created"];
	        this.length = source["length"];
	        this.video_review = source["video_review"];
	        this.aid = source["aid"];
	        this.bvid = source["bvid"];
	        this.hide_click = source["hide_click"];
	        this.is_pay = source["is_pay"];
	        this.is_union_video = source["is_union_video"];
	        this.is_steins_gate = source["is_steins_gate"];
	        this.is_live_playback = source["is_live_playback"];
	        this.is_lesson_video = source["is_lesson_video"];
	        this.is_lesson_finished = source["is_lesson_finished"];
	        this.lesson_update_info = source["lesson_update_info"];
	        this.jump_url = source["jump_url"];
	        this.meta = this.convertValues(source["meta"], VideoVListItemMeta);
	        this.is_avoided = source["is_avoided"];
	        this.season_id = source["season_id"];
	        this.attribute = source["attribute"];
	        this.is_charging_arc = source["is_charging_arc"];
	        this.elec_arc_type = source["elec_arc_type"];
	        this.elec_arc_badge = source["elec_arc_badge"];
	        this.vt = source["vt"];
	        this.enable_vt = source["enable_vt"];
	        this.vt_display = source["vt_display"];
	        this.playback_position = source["playback_position"];
	        this.is_self_view = source["is_self_view"];
	        this.view_self_type = source["view_self_type"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class VideoDescV2 {
	    raw_text: string;
	    type: number;
	    biz_id: number;
	
	    static createFrom(source: any = {}) {
	        return new VideoDescV2(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.raw_text = source["raw_text"];
	        this.type = source["type"];
	        this.biz_id = source["biz_id"];
	    }
	}
	export class VideoDetailArgueInfo {
	    argue_msg: string;
	    argue_type: number;
	    argue_link: string;
	
	    static createFrom(source: any = {}) {
	        return new VideoDetailArgueInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.argue_msg = source["argue_msg"];
	        this.argue_type = source["argue_type"];
	        this.argue_link = source["argue_link"];
	    }
	}
	export class VideoSubTitle {
	    allow_submit: boolean;
	    list: any[];
	
	    static createFrom(source: any = {}) {
	        return new VideoSubTitle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.allow_submit = source["allow_submit"];
	        this.list = source["list"];
	    }
	}
	export class VideoViewStat {
	    aid: number;
	    view: number;
	    danmaku: number;
	    reply: number;
	    favorite: number;
	    coin: number;
	    share: number;
	    now_rank: number;
	    his_rank: number;
	    like: number;
	    dislike: number;
	    evaluation: string;
	    vt: number;
	
	    static createFrom(source: any = {}) {
	        return new VideoViewStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.aid = source["aid"];
	        this.view = source["view"];
	        this.danmaku = source["danmaku"];
	        this.reply = source["reply"];
	        this.favorite = source["favorite"];
	        this.coin = source["coin"];
	        this.share = source["share"];
	        this.now_rank = source["now_rank"];
	        this.his_rank = source["his_rank"];
	        this.like = source["like"];
	        this.dislike = source["dislike"];
	        this.evaluation = source["evaluation"];
	        this.vt = source["vt"];
	    }
	}
	export class VideoViewRights {
	    bp: number;
	    elec: number;
	    download: number;
	    movie: number;
	    pay: number;
	    hd5: number;
	    no_reprint: number;
	    autoplay: number;
	    ugc_pay: number;
	    is_cooperation: number;
	    ugc_pay_preview: number;
	    no_background: number;
	    clean_mode: number;
	    is_stein_gate: number;
	    is_360: number;
	    no_share: number;
	    arc_pay: number;
	    free_watch: number;
	
	    static createFrom(source: any = {}) {
	        return new VideoViewRights(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bp = source["bp"];
	        this.elec = source["elec"];
	        this.download = source["download"];
	        this.movie = source["movie"];
	        this.pay = source["pay"];
	        this.hd5 = source["hd5"];
	        this.no_reprint = source["no_reprint"];
	        this.autoplay = source["autoplay"];
	        this.ugc_pay = source["ugc_pay"];
	        this.is_cooperation = source["is_cooperation"];
	        this.ugc_pay_preview = source["ugc_pay_preview"];
	        this.no_background = source["no_background"];
	        this.clean_mode = source["clean_mode"];
	        this.is_stein_gate = source["is_stein_gate"];
	        this.is_360 = source["is_360"];
	        this.no_share = source["no_share"];
	        this.arc_pay = source["arc_pay"];
	        this.free_watch = source["free_watch"];
	    }
	}
	export class VideoPage {
	    cid: number;
	    page: number;
	    from: string;
	    part: string;
	    duration: number;
	    vid: string;
	    weblink: string;
	    dimension: Dimension;
	    first_frame: string;
	    ctime: number;
	
	    static createFrom(source: any = {}) {
	        return new VideoPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cid = source["cid"];
	        this.page = source["page"];
	        this.from = source["from"];
	        this.part = source["part"];
	        this.duration = source["duration"];
	        this.vid = source["vid"];
	        this.weblink = source["weblink"];
	        this.dimension = this.convertValues(source["dimension"], Dimension);
	        this.first_frame = source["first_frame"];
	        this.ctime = source["ctime"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class VideoDetailView {
	    pages: VideoPage[];
	    cid: number;
	    bvid: string;
	    aid: number;
	    duration: number;
	    videos: number;
	    ugc_season: UgcSeason;
	    pic: string;
	    title: string;
	    pubdate: number;
	    ctime: number;
	    tid: number;
	    tid_v2: number;
	    tname: string;
	    tname_v2: string;
	    copyright: number;
	    desc: string;
	    desc_v2: VideoDescV2[];
	    state: number;
	    mission_id: number;
	    rights: VideoViewRights;
	    owner: VideoOwner;
	    stat: VideoViewStat;
	    argue_info: VideoDetailArgueInfo;
	    dynamic: string;
	    dimension: Dimension;
	    season_id: number;
	    premiere: any;
	    teenage_mode: number;
	    is_chargeable_season: boolean;
	    is_story: boolean;
	    is_upower_exclusive: boolean;
	    is_upower_play: boolean;
	    is_upower_preview: boolean;
	    enable_vt: number;
	    vt_display: string;
	    is_upower_exclusive_with_qa: boolean;
	    no_cache: boolean;
	    subtitle: VideoSubTitle;
	    is_season_display: boolean;
	    user_garb: UserGarb;
	    honor_reply: any;
	    like_icon: string;
	    need_jump_bv: boolean;
	    disable_show_up_info: boolean;
	    is_story_play: number;
	    is_view_self: boolean;
	
	    static createFrom(source: any = {}) {
	        return new VideoDetailView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pages = this.convertValues(source["pages"], VideoPage);
	        this.cid = source["cid"];
	        this.bvid = source["bvid"];
	        this.aid = source["aid"];
	        this.duration = source["duration"];
	        this.videos = source["videos"];
	        this.ugc_season = this.convertValues(source["ugc_season"], UgcSeason);
	        this.pic = source["pic"];
	        this.title = source["title"];
	        this.pubdate = source["pubdate"];
	        this.ctime = source["ctime"];
	        this.tid = source["tid"];
	        this.tid_v2 = source["tid_v2"];
	        this.tname = source["tname"];
	        this.tname_v2 = source["tname_v2"];
	        this.copyright = source["copyright"];
	        this.desc = source["desc"];
	        this.desc_v2 = this.convertValues(source["desc_v2"], VideoDescV2);
	        this.state = source["state"];
	        this.mission_id = source["mission_id"];
	        this.rights = this.convertValues(source["rights"], VideoViewRights);
	        this.owner = this.convertValues(source["owner"], VideoOwner);
	        this.stat = this.convertValues(source["stat"], VideoViewStat);
	        this.argue_info = this.convertValues(source["argue_info"], VideoDetailArgueInfo);
	        this.dynamic = source["dynamic"];
	        this.dimension = this.convertValues(source["dimension"], Dimension);
	        this.season_id = source["season_id"];
	        this.premiere = source["premiere"];
	        this.teenage_mode = source["teenage_mode"];
	        this.is_chargeable_season = source["is_chargeable_season"];
	        this.is_story = source["is_story"];
	        this.is_upower_exclusive = source["is_upower_exclusive"];
	        this.is_upower_play = source["is_upower_play"];
	        this.is_upower_preview = source["is_upower_preview"];
	        this.enable_vt = source["enable_vt"];
	        this.vt_display = source["vt_display"];
	        this.is_upower_exclusive_with_qa = source["is_upower_exclusive_with_qa"];
	        this.no_cache = source["no_cache"];
	        this.subtitle = this.convertValues(source["subtitle"], VideoSubTitle);
	        this.is_season_display = source["is_season_display"];
	        this.user_garb = this.convertValues(source["user_garb"], UserGarb);
	        this.honor_reply = source["honor_reply"];
	        this.like_icon = source["like_icon"];
	        this.need_jump_bv = source["need_jump_bv"];
	        this.disable_show_up_info = source["disable_show_up_info"];
	        this.is_story_play = source["is_story_play"];
	        this.is_view_self = source["is_view_self"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class VideoDetailConciseData {
	    view: VideoDetailView;
	
	    static createFrom(source: any = {}) {
	        return new VideoDetailConciseData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.view = this.convertValues(source["view"], VideoDetailView);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class VideoList {
	    slist: any[];
	    tlist: any;
	    vlist: VListItem[];
	
	    static createFrom(source: any = {}) {
	        return new VideoList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.slist = source["slist"];
	        this.tlist = source["tlist"];
	        this.vlist = this.convertValues(source["vlist"], VListItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class VideoListPage {
	    pn: number;
	    ps: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new VideoListPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pn = source["pn"];
	        this.ps = source["ps"];
	        this.count = source["count"];
	    }
	}
	export class VideoListData {
	    list: VideoList;
	    page: VideoListPage;
	    episodic_button: EpisodicButton;
	    is_risk: boolean;
	    gaia_res_type: number;
	    gaia_data: any;
	
	    static createFrom(source: any = {}) {
	        return new VideoListData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.list = this.convertValues(source["list"], VideoList);
	        this.page = this.convertValues(source["page"], VideoListPage);
	        this.episodic_button = this.convertValues(source["episodic_button"], EpisodicButton);
	        this.is_risk = source["is_risk"];
	        this.gaia_res_type = source["gaia_res_type"];
	        this.gaia_data = source["gaia_data"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	
	
	export class VolumeMultiSceneArgs {
	    high_dynamic_target_i: string;
	    normal_target_i: string;
	    undersized_target_i: string;
	
	    static createFrom(source: any = {}) {
	        return new VolumeMultiSceneArgs(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.high_dynamic_target_i = source["high_dynamic_target_i"];
	        this.normal_target_i = source["normal_target_i"];
	        this.undersized_target_i = source["undersized_target_i"];
	    }
	}
	export class Volume {
	    measured_i: number;
	    measured_lra: number;
	    measured_tp: number;
	    measured_threshold: number;
	    target_offset: number;
	    target_i: number;
	    target_tp: number;
	    multi_scene_args: VolumeMultiSceneArgs;
	
	    static createFrom(source: any = {}) {
	        return new Volume(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.measured_i = source["measured_i"];
	        this.measured_lra = source["measured_lra"];
	        this.measured_tp = source["measured_tp"];
	        this.measured_threshold = source["measured_threshold"];
	        this.target_offset = source["target_offset"];
	        this.target_i = source["target_i"];
	        this.target_tp = source["target_tp"];
	        this.multi_scene_args = this.convertValues(source["multi_scene_args"], VolumeMultiSceneArgs);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class VideoURLData {
	    from: string;
	    result: string;
	    message: string;
	    quality: number;
	    format: string;
	    timelength: number;
	    accept_format: string;
	    accept_description: string[];
	    accept_quality: number[];
	    video_codecid: number;
	    seek_param: string;
	    seek_type: string;
	    dash: Dash;
	    support_formats: SupportFormat[];
	    high_format: any;
	    volume: Volume;
	    last_play_time: number;
	    last_play_cid: number;
	    view_info: any;
	    play_conf: PlayConf;
	    cur_language: string;
	    cur_production_type: number;
	    auto_qn_resp: AutoQnResp;
	
	    static createFrom(source: any = {}) {
	        return new VideoURLData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.result = source["result"];
	        this.message = source["message"];
	        this.quality = source["quality"];
	        this.format = source["format"];
	        this.timelength = source["timelength"];
	        this.accept_format = source["accept_format"];
	        this.accept_description = source["accept_description"];
	        this.accept_quality = source["accept_quality"];
	        this.video_codecid = source["video_codecid"];
	        this.seek_param = source["seek_param"];
	        this.seek_type = source["seek_type"];
	        this.dash = this.convertValues(source["dash"], Dash);
	        this.support_formats = this.convertValues(source["support_formats"], SupportFormat);
	        this.high_format = source["high_format"];
	        this.volume = this.convertValues(source["volume"], Volume);
	        this.last_play_time = source["last_play_time"];
	        this.last_play_cid = source["last_play_cid"];
	        this.view_info = source["view_info"];
	        this.play_conf = this.convertValues(source["play_conf"], PlayConf);
	        this.cur_language = source["cur_language"];
	        this.cur_production_type = source["cur_production_type"];
	        this.auto_qn_resp = this.convertValues(source["auto_qn_resp"], AutoQnResp);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	

}

export namespace pb {
	
	export class Match {
	    prefix?: number[];
	    ignore_bytes?: string;
	
	    static createFrom(source: any = {}) {
	        return new Match(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prefix = source["prefix"];
	        this.ignore_bytes = source["ignore_bytes"];
	    }
	}

}

export namespace ristretto {
	
	export class Metrics {
	
	
	    static createFrom(source: any = {}) {
	        return new Metrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

