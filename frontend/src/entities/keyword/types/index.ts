// Keyword 엔티티 타입
import * as models from '../../../../wailsjs/go/models';

export type KeywordAnalysisResult = models.main.KeywordAnalysisResult;
export type RelatedDetail = models.main.RelatedDetail;
export type AutoExtractResult = models.main.AutoExtractResult;
export type Usage = models.main.Usage;
export type CreatorAdvisorResponse = models.main.CreatorAdvisorResponse;
export type SearchAdTestResult = models.main.SearchAdTestResult;

export interface CreatorAdvisorPayload {
  loginId: string;
  password: string;
  showBrowser?: boolean;
}

// Progress는 shared에서 정의
export type { Progress } from '../../../shared/api/wails';
