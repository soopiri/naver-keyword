// API wrapper for Wails backend
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';
import * as App from '../wailsjs/go/main/App';
import * as models from '../wailsjs/go/models';

// Re-export types from Wails generated models
export type Config = models.main.Config;
export type Usage = models.main.Usage;
export type KeywordAnalysisResult = models.main.KeywordAnalysisResult;
export type RelatedDetail = models.main.RelatedDetail;
export type AutoExtractResult = models.main.AutoExtractResult;
export type CreatorAdvisorResponse = models.main.CreatorAdvisorResponse;
export type SearchAdTestResult = models.main.SearchAdTestResult;

// Local types for payloads and progress
export interface CreatorAdvisorPayload {
  loginId: string;
  password: string;
  showBrowser?: boolean;
}

export interface Progress {
  current: number;
  total: number;
  message?: string;
  detail?: string;
  keyword?: string;
  stage?: string;
  step?: string;
  jobId?: string;
}

// API functions using Wails generated bindings
export const api = {
  getConfig: (): Promise<Config> => {
    return App.GetConfig();
  },

  saveConfig: (config: Config): Promise<void> => {
    // Convert Config class instance to plain object for proper serialization
    const plainConfig = {
      naver: {
        clientId: config.naver.clientId,
        clientSecret: config.naver.clientSecret,
      },
      creatorAdvisor: {
        loginId: config.creatorAdvisor.loginId,
        password: config.creatorAdvisor.password,
      },
      searchad: {
        baseUrl: config.searchad.baseUrl,
        customerId: config.searchad.customerId,
        accessKey: config.searchad.accessKey,
        secretKey: config.searchad.secretKey,
      },
      scoring: {
        w_pc: config.scoring.w_pc,
        w_mob: config.scoring.w_mob,
        k: config.scoring.k,
        alpha: config.scoring.alpha,
        t: config.scoring.t,
        r: config.scoring.r,
      },
      seeds: config.seeds,
      usage: {
        date: config.usage.date,
        used: config.usage.used,
      },
    };
    return App.SaveConfig(plainConfig as any);
  },

  getUsage: (): Promise<Usage> => {
    return App.GetUsage();
  },

  analyzeKeyword: (keyword: string): Promise<KeywordAnalysisResult> => {
    return App.AnalyzeKeyword(keyword);
  },

  runAutoExtract: (payload: { seeds?: string[] }): Promise<AutoExtractResult[]> => {
    return App.RunAutoExtract(payload);
  },

  collectCreatorAdvisor: (payload: CreatorAdvisorPayload): Promise<CreatorAdvisorResponse> => {
    return App.CollectCreatorAdvisor(JSON.stringify(payload));
  },

  testSearchAd: (): Promise<SearchAdTestResult> => {
    return App.TestSearchAd();
  },

  getSearchAdLog: (): Promise<string> => {
    return App.GetSearchAdLog();
  },

  // Event listeners
  onAutoProgress: (callback: (progress: Progress) => void): (() => void) => {
    const unsubscribe = EventsOn('analysis:auto:progress', callback);
    return () => {
      EventsOff('analysis:auto:progress');
      if (unsubscribe) unsubscribe();
    };
  },

  onKeywordProgress: (callback: (progress: Progress) => void): (() => void) => {
    const unsubscribe = EventsOn('analysis:keyword:progress', callback);
    return () => {
      EventsOff('analysis:keyword:progress');
      if (unsubscribe) unsubscribe();
    };
  },

  onCreatorProgress: (callback: (progress: Progress) => void): (() => void) => {
    const unsubscribe = EventsOn('creator:progress', callback);
    return () => {
      EventsOff('creator:progress');
      if (unsubscribe) unsubscribe();
    };
  },

  onCreatorDone: (callback: (data: { jobId: string; result: any }) => void): (() => void) => {
    const unsubscribe = EventsOn('creator:done', callback);
    return () => {
      EventsOff('creator:done');
      if (unsubscribe) unsubscribe();
    };
  },

  onCreatorError: (callback: (error: { jobId: string; message: string }) => void): (() => void) => {
    const unsubscribe = EventsOn('creator:error', callback);
    return () => {
      EventsOff('creator:error');
      if (unsubscribe) unsubscribe();
    };
  },
};
