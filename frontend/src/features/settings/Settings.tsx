// Settings Feature - 완결된 UI 제공
import { useState } from "react";
import { configApi } from "../../entities/config";
import { keywordApi } from "../../entities/keyword";
import type { Config } from "../../entities/config";
import * as models from "../../../wailsjs/go/models";

interface SettingsProps {
  config: Config;
  onSave: (config: Config) => void;
}

export function Settings({ config, onSave }: SettingsProps) {
  const [localConfig, setLocalConfig] = useState<Config>(config);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<string>("");

  const handleChange = (
    section: keyof Config,
    field: string,
    value: string,
  ) => {
    setLocalConfig((prev: Config) => {
      const newConfig = models.main.Config.createFrom(prev);
      const sectionObj = (newConfig as any)[section];
      if (sectionObj) {
        (sectionObj as any)[field] = value;
      }
      return newConfig;
    });
  };

  const handleScoringChange = (field: string, value: number) => {
    setLocalConfig((prev: Config) => {
      const newConfig = models.main.Config.createFrom(prev);
      (newConfig.scoring as any)[field] = value;
      return newConfig;
    });
  };

  const handleTest = async () => {
    setTesting(true);
    setTestResult("");
    try {
      const result = await keywordApi.testSearchAd();
      if (result.ok) {
        setTestResult("✅ 모든 API 연결이 정상입니다.");
      } else {
        setTestResult(
          `❌ 연결 실패:\nSearchAd: ${result.searchad.message}\nNaver: ${result.naver.message}`,
        );
      }
    } catch (error: any) {
      setTestResult(`❌ 테스트 실패: ${error.message}`);
    } finally {
      setTesting(false);
    }
  };

  return (
    <>
      <section className="card">
        <h2 className="title-chip">네이버 대표 계정</h2>
        <div className="grid">
          <label>
            네이버 ID
            <input
              value={localConfig.creatorAdvisor?.loginId || ""}
              onChange={(e) =>
                handleChange("creatorAdvisor", "loginId", e.target.value)
              }
            />
          </label>
          <label>
            네이버 PW
            <div className="input-row">
              <input
                type="password"
                value={localConfig.creatorAdvisor?.password || ""}
                onChange={(e) =>
                  handleChange("creatorAdvisor", "password", e.target.value)
                }
              />
            </div>
          </label>
        </div>
      </section>

      <section className="card">
        <h2 className="title-chip">API 설정</h2>
        <div className="grid">
          <label>
            네이버 Client ID
            <input
              value={localConfig.naver.clientId}
              onChange={(e) =>
                handleChange("naver", "clientId", e.target.value)
              }
            />
          </label>
          <label>
            네이버 Client Secret
            <input
              value={localConfig.naver.clientSecret}
              onChange={(e) =>
                handleChange("naver", "clientSecret", e.target.value)
              }
            />
          </label>
          <label>
            Customer ID (SearchAd)
            <input
              value={localConfig.searchad.customerId}
              onChange={(e) =>
                handleChange("searchad", "customerId", e.target.value)
              }
            />
          </label>
          <label>
            엑세스 라이선스
            <input
              value={localConfig.searchad.accessKey}
              onChange={(e) =>
                handleChange("searchad", "accessKey", e.target.value)
              }
            />
          </label>
          <label>
            Secret Key
            <input
              value={localConfig.searchad.secretKey}
              onChange={(e) =>
                handleChange("searchad", "secretKey", e.target.value)
              }
            />
          </label>
        </div>
        <div className="row">
          <button
            className="btn btn-primary"
            onClick={() => onSave(localConfig)}
          >
            저장
          </button>
          <button className="btn" onClick={handleTest} disabled={testing}>
            API 연결 테스트
          </button>
          <span className="status">{testResult}</span>
        </div>
        {testResult && (
          <div className="result">
            <div className="row">
              <div>
                <strong>연결 상태:</strong>{" "}
                {testResult.includes("✅") ? "성공" : "실패"}
              </div>
            </div>
          </div>
        )}
      </section>

      <section className="card">
        <div className="row between">
          <div className="row">
            <h2>
              <span className="title-chip">스코어링 파라미터</span>
            </h2>
            <span className="muted inline-note">
              키워드 점수 계산에 사용됩니다.
            </span>
          </div>
        </div>
        <div className="grid">
          <label>
            w_pc
            <input
              type="number"
              step="0.1"
              value={localConfig.scoring.w_pc}
              onChange={(e) =>
                handleScoringChange("w_pc", parseFloat(e.target.value))
              }
            />
          </label>
          <label>
            w_mob
            <input
              type="number"
              step="0.1"
              value={localConfig.scoring.w_mob}
              onChange={(e) =>
                handleScoringChange("w_mob", parseFloat(e.target.value))
              }
            />
          </label>
          <label>
            k
            <input
              type="number"
              step="0.1"
              value={localConfig.scoring.k}
              onChange={(e) =>
                handleScoringChange("k", parseFloat(e.target.value))
              }
            />
          </label>
          <label>
            alpha
            <input
              type="number"
              step="0.1"
              value={localConfig.scoring.alpha}
              onChange={(e) =>
                handleScoringChange("alpha", parseFloat(e.target.value))
              }
            />
          </label>
          <label>
            t
            <input
              type="number"
              step="0.1"
              value={localConfig.scoring.t}
              onChange={(e) =>
                handleScoringChange("t", parseFloat(e.target.value))
              }
            />
          </label>
          <label>
            r
            <input
              type="number"
              step="0.1"
              value={localConfig.scoring.r}
              onChange={(e) =>
                handleScoringChange("r", parseFloat(e.target.value))
              }
            />
          </label>
        </div>
      </section>
    </>
  );
}
