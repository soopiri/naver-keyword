// Config 엔티티 API
import { App } from '../../../shared/api/wails';
import type { Config } from '../types';

export const configApi = {
  get: (): Promise<Config> => {
    return App.GetConfig();
  },

  save: (config: Config): Promise<void> => {
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
};
