export namespace model {
	
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

}

