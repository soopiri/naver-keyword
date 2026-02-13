// Keyword 엔티티 API
import { App } from '../../../shared/api/wails';
import type {
  KeywordAnalysisResult,
  AutoExtractResult,
  Usage,
  CreatorAdvisorResponse,
  SearchAdTestResult,
  CreatorAdvisorPayload,
} from '../types';

export const keywordApi = {
  analyze: (keyword: string): Promise<KeywordAnalysisResult> => {
    return App.AnalyzeKeyword(keyword);
  },

  runAutoExtract: (payload: { seeds?: string[] }): Promise<AutoExtractResult[]> => {
    return App.RunAutoExtract(payload);
  },

  getUsage: (): Promise<Usage> => {
    return App.GetUsage();
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
};
