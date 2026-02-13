// Wails 공통 유틸리티
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';
import * as App from '../../../wailsjs/go/main/App';
import * as models from '../../../wailsjs/go/models';

// Progress 타입 (공통)
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

// Wails 이벤트 헬퍼
export const wailsEvents = {
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

// Wails App 바인딩 재export
export { App, models };
