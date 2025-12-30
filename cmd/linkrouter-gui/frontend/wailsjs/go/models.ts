export namespace config {
	
	export class Rule {
	    regex: string;
	    program: string;
	    arguments: string;
	
	    static createFrom(source: any = {}) {
	        return new Rule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.regex = source["regex"];
	        this.program = source["program"];
	        this.arguments = source["arguments"];
	    }
	}
	export class GlobalConfig {
	    fallbackBrowserPath: string;
	    fallbackBrowserArgs: string;
	    defaultConfigEditor: string;
	    logPath: string;
	    supportedProtocols: string[];
	
	    static createFrom(source: any = {}) {
	        return new GlobalConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fallbackBrowserPath = source["fallbackBrowserPath"];
	        this.fallbackBrowserArgs = source["fallbackBrowserArgs"];
	        this.defaultConfigEditor = source["defaultConfigEditor"];
	        this.logPath = source["logPath"];
	        this.supportedProtocols = source["supportedProtocols"];
	    }
	}
	export class Config {
	    global: GlobalConfig;
	    rules: Rule[];
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.global = this.convertValues(source["global"], GlobalConfig);
	        this.rules = this.convertValues(source["rules"], Rule);
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

