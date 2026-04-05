export namespace model {
	
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
	export class MyInfoProfileVip {
	    type: number;
	    status: number;
	    due_date: number;
	    theme_type: number;
	    label: MyInfoVipLabel;
	    avatar_subscript: number;
	    nickname_color: string;
	    OttInfo: MyInfoVipOttInfo;
	    super_vip: MyInfoVipSuperVip;
	
	    static createFrom(source: any = {}) {
	        return new MyInfoProfileVip(source);
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
	    vip: MyInfoProfileVip;
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
	        this.vip = this.convertValues(source["vip"], MyInfoProfileVip);
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
	
	
	

}

