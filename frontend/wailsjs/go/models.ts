export namespace main {
	
	export class AutoExtractResult {
	    keyword: string;
	    searchPc: number;
	    searchMobile: number;
	    totalSearch: number;
	    monthlySearch?: number;
	    monthlyAvgClicks?: number;
	    monthlyAvgCtr?: number;
	    competition: any;
	    docCount: number;
	    group: string;
	    score: number;
	    cpc?: number;
	    cpcPc?: number;
	    cpcMobile?: number;
	    pcClicks: number;
	    pcCost: number;
	    mobileClicks: number;
	    mobileCost: number;
	    totalClicks: number;
	    totalCost: number;
	
	    static createFrom(source: any = {}) {
	        return new AutoExtractResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.keyword = source["keyword"];
	        this.searchPc = source["searchPc"];
	        this.searchMobile = source["searchMobile"];
	        this.totalSearch = source["totalSearch"];
	        this.monthlySearch = source["monthlySearch"];
	        this.monthlyAvgClicks = source["monthlyAvgClicks"];
	        this.monthlyAvgCtr = source["monthlyAvgCtr"];
	        this.competition = source["competition"];
	        this.docCount = source["docCount"];
	        this.group = source["group"];
	        this.score = source["score"];
	        this.cpc = source["cpc"];
	        this.cpcPc = source["cpcPc"];
	        this.cpcMobile = source["cpcMobile"];
	        this.pcClicks = source["pcClicks"];
	        this.pcCost = source["pcCost"];
	        this.mobileClicks = source["mobileClicks"];
	        this.mobileCost = source["mobileCost"];
	        this.totalClicks = source["totalClicks"];
	        this.totalCost = source["totalCost"];
	    }
	}
	export class Usage {
	    date: string;
	    used: number;
	
	    static createFrom(source: any = {}) {
	        return new Usage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.used = source["used"];
	    }
	}
	export class Scoring {
	    w_pc: number;
	    w_mob: number;
	    k: number;
	    alpha: number;
	    t: number;
	    r: number;
	
	    static createFrom(source: any = {}) {
	        return new Scoring(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.w_pc = source["w_pc"];
	        this.w_mob = source["w_mob"];
	        this.k = source["k"];
	        this.alpha = source["alpha"];
	        this.t = source["t"];
	        this.r = source["r"];
	    }
	}
	export class SearchAdConfig {
	    baseUrl: string;
	    customerId: string;
	    accessKey: string;
	    secretKey: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchAdConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.baseUrl = source["baseUrl"];
	        this.customerId = source["customerId"];
	        this.accessKey = source["accessKey"];
	        this.secretKey = source["secretKey"];
	    }
	}
	export class CreatorAdvisorConfig {
	    loginId: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new CreatorAdvisorConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.loginId = source["loginId"];
	        this.password = source["password"];
	    }
	}
	export class NaverConfig {
	    clientId: string;
	    clientSecret: string;
	
	    static createFrom(source: any = {}) {
	        return new NaverConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.clientId = source["clientId"];
	        this.clientSecret = source["clientSecret"];
	    }
	}
	export class Config {
	    naver: NaverConfig;
	    creatorAdvisor: CreatorAdvisorConfig;
	    searchad: SearchAdConfig;
	    scoring: Scoring;
	    seeds: string[];
	    usage: Usage;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.naver = this.convertValues(source["naver"], NaverConfig);
	        this.creatorAdvisor = this.convertValues(source["creatorAdvisor"], CreatorAdvisorConfig);
	        this.searchad = this.convertValues(source["searchad"], SearchAdConfig);
	        this.scoring = this.convertValues(source["scoring"], Scoring);
	        this.seeds = source["seeds"];
	        this.usage = this.convertValues(source["usage"], Usage);
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
	
	export class CreatorAdvisorResponse {
	    ok: boolean;
	    jobId?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new CreatorAdvisorResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.jobId = source["jobId"];
	        this.message = source["message"];
	    }
	}
	export class RelatedDetail {
	    keyword: string;
	    searchPc: number;
	    searchMobile: number;
	    totalSearch: number;
	    docCount: number;
	
	    static createFrom(source: any = {}) {
	        return new RelatedDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.keyword = source["keyword"];
	        this.searchPc = source["searchPc"];
	        this.searchMobile = source["searchMobile"];
	        this.totalSearch = source["totalSearch"];
	        this.docCount = source["docCount"];
	    }
	}
	export class KeywordAnalysisResult {
	    keyword: string;
	    searchPc: number;
	    searchMobile: number;
	    totalSearch: number;
	    monthlySearch?: number;
	    monthlyAvgClicks?: number;
	    monthlyAvgCtr?: number;
	    competition: any;
	    docCount: number;
	    group: string;
	    relatedDetails: RelatedDetail[];
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new KeywordAnalysisResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.keyword = source["keyword"];
	        this.searchPc = source["searchPc"];
	        this.searchMobile = source["searchMobile"];
	        this.totalSearch = source["totalSearch"];
	        this.monthlySearch = source["monthlySearch"];
	        this.monthlyAvgClicks = source["monthlyAvgClicks"];
	        this.monthlyAvgCtr = source["monthlyAvgCtr"];
	        this.competition = source["competition"];
	        this.docCount = source["docCount"];
	        this.group = source["group"];
	        this.relatedDetails = this.convertValues(source["relatedDetails"], RelatedDetail);
	        this.score = source["score"];
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
	
	
	
	
	export class TestResult {
	    ok: boolean;
	    status: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new TestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.status = source["status"];
	        this.message = source["message"];
	    }
	}
	export class SearchAdTestResult {
	    ok: boolean;
	    searchad: TestResult;
	    naver: TestResult;
	
	    static createFrom(source: any = {}) {
	        return new SearchAdTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.searchad = this.convertValues(source["searchad"], TestResult);
	        this.naver = this.convertValues(source["naver"], TestResult);
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

