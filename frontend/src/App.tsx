import { useState, useEffect, useRef } from "react";
import "./style.css";
import { api, type Config, type Progress } from "./api";
import * as models from "../wailsjs/go/models";

function App() {
  const [showSettings, setShowSettings] = useState(false);
  const [config, setConfig] = useState<Config | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      const loadedConfig = await api.getConfig();
      setConfig(loadedConfig);
    } catch (error) {
      console.error("설정 로드 실패:", error);
    } finally {
      setLoading(false);
    }
  };

  const saveConfig = async (newConfig: Config) => {
    try {
      await api.saveConfig(newConfig);
      setConfig(newConfig);
      alert("설정이 저장되었습니다.");
    } catch (error) {
      console.error("설정 저장 실패:", error);
      alert("설정 저장에 실패했습니다.");
    }
  };

  if (loading) {
    return (
      <div className="app">
        <div className="loading">로딩 중...</div>
      </div>
    );
  }

  return (
    <div className="app">
      <header className="header">
        <div className="title-block">
          <div>
            <h1>네이버 키워드 수집 프로그램</h1>
            <p>네이버 API를 통해서 키워드 데이터를 수집합니다.</p>
          </div>
        </div>
        <button
          className="btn btn-info"
          onClick={() => setShowSettings(!showSettings)}
        >
          설정
        </button>
      </header>

      {showSettings && config && (
        <SettingsPage config={config} onSave={saveConfig} />
      )}

      <GoldenKeywordSearchPage />
      <TrendDataSearchPage />
      <SingleKeywordSearchPage />
    </div>
  );
}

// Settings Page
function SettingsPage({
  config,
  onSave,
}: {
  config: Config;
  onSave: (config: Config) => void;
}) {
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
      const result = await api.testSearchAd();
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

// Analyze Page (not used in main UI, but kept for reference)
function AnalyzePage() {
  const [keyword, setKeyword] = useState("");
  const [analyzing, setAnalyzing] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [progress, setProgress] = useState<any>(null);

  useEffect(() => {
    const unsubscribe = api.onKeywordProgress((progress: Progress) => {
      setProgress(progress);
    });
    return () => unsubscribe();
  }, []);

  const handleAnalyze = async () => {
    if (!keyword.trim()) {
      alert("키워드를 입력하세요.");
      return;
    }

    setAnalyzing(true);
    setResult(null);
    setProgress(null);

    try {
      const analysisResult = await api.analyzeKeyword(keyword);
      setResult(analysisResult);
    } catch (error: any) {
      alert(`분석 실패: ${error.message}`);
    } finally {
      setAnalyzing(false);
    }
  };

  return (
    <div className="analyze-page">
      <h2>키워드 분석</h2>
      <div className="input-section">
        <input
          type="text"
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          placeholder="분석할 키워드를 입력하세요"
          onKeyPress={(e) => e.key === "Enter" && handleAnalyze()}
          disabled={analyzing}
        />
        <button onClick={handleAnalyze} disabled={analyzing || !keyword.trim()}>
          {analyzing ? "분석 중..." : "분석 시작"}
        </button>
      </div>

      {progress && (
        <div className="progress-section">
          <div className="progress-bar">
            <div
              className="progress-fill"
              style={{ width: `${(progress.current / progress.total) * 100}%` }}
            />
          </div>
          <div className="progress-text">
            {progress.current}/{progress.total} -{" "}
            {progress.message || progress.stage}
            {progress.keyword && ` (${progress.keyword})`}
          </div>
        </div>
      )}

      {result && (
        <div className="result-section">
          <h3>분석 결과</h3>
          <div className="result-card">
            <div className="result-row">
              <strong>키워드:</strong> {result.keyword}
            </div>
            <div className="result-row">
              <strong>검색량:</strong> PC {result.searchPc.toLocaleString()} /
              모바일 {result.searchMobile.toLocaleString()} (총{" "}
              {result.totalSearch.toLocaleString()})
            </div>
            <div className="result-row">
              <strong>문서 수:</strong> {result.docCount.toLocaleString()}
            </div>
            <div className="result-row">
              <strong>그룹:</strong> {result.group}
            </div>
            <div className="result-row">
              <strong>점수:</strong> {result.score.toFixed(2)}
            </div>
            {result.monthlySearch && (
              <div className="result-row">
                <strong>월간 검색량:</strong>{" "}
                {result.monthlySearch.toLocaleString()}
              </div>
            )}
            {result.competition && (
              <div className="result-row">
                <strong>경쟁도:</strong> {result.competition}
              </div>
            )}

            {result.relatedDetails && result.relatedDetails.length > 0 && (
              <div className="related-section">
                <h4>관련 키워드</h4>
                <table className="result-table">
                  <thead>
                    <tr>
                      <th>키워드</th>
                      <th>PC 검색량</th>
                      <th>모바일 검색량</th>
                      <th>총 검색량</th>
                      <th>문서 수</th>
                    </tr>
                  </thead>
                  <tbody>
                    {result.relatedDetails.map((item: any, idx: number) => (
                      <tr key={idx}>
                        <td>{item.keyword}</td>
                        <td>{item.searchPc.toLocaleString()}</td>
                        <td>{item.searchMobile.toLocaleString()}</td>
                        <td>{item.totalSearch.toLocaleString()}</td>
                        <td>{item.docCount.toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

// 황금 키워드 검색 Page
function GoldenKeywordSearchPage() {
  const [seedInput, setSeedInput] = useState("");
  const [extracting, setExtracting] = useState(false);
  const [results, setResults] = useState<any[]>([]);
  const [progress, setProgress] = useState<any>(null);
  const [usage, setUsage] = useState(0);
  const [relatedKeyword, setRelatedKeyword] = useState<string>("");
  const [relatedLoading, setRelatedLoading] = useState(false);
  const [relatedItems, setRelatedItems] = useState<any[]>([]);
  const [relatedStatus, setRelatedStatus] = useState<Record<string, string>>(
    {},
  );
  const extractingRef = useRef(false);
  const [sortField, setSortField] = useState<string | null>(null);
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");
  const [infoModal, setInfoModal] = useState<{ title: string; description: string } | null>(null);
  const [showScoreInfo, setShowScoreInfo] = useState(false);

  useEffect(() => {
    const unsubscribe = api.onAutoProgress((progress: Progress) => {
      setProgress(progress);
    });
    return () => unsubscribe();
  }, []);

  useEffect(() => {
    loadUsage();
  }, []);

  const loadUsage = async () => {
    try {
      const usageData = await api.getUsage();
      setUsage(usageData.used || 0);
    } catch (error) {
      console.error("사용량 로드 실패:", error);
    }
  };

  const handleReset = () => {
    setSeedInput("");
    setResults([]);
    setProgress(null);
    setRelatedKeyword("");
    setRelatedItems([]);
    setRelatedStatus({});
  };

  const handleAnalyzeRelated = async (keyword: string) => {
    if (!keyword) return;

    setRelatedStatus((prev) => ({ ...prev, [keyword]: "loading" }));
    setRelatedKeyword(keyword);
    setRelatedLoading(true);
    setRelatedItems([]);

    try {
      const result = await api.analyzeKeyword(keyword);
      setRelatedItems(result.relatedDetails || []);
      setRelatedStatus((prev) => ({ ...prev, [keyword]: "done" }));
    } catch (error: any) {
      setRelatedStatus((prev) => ({ ...prev, [keyword]: "error" }));
      alert(`연관검색어 수집 실패: ${error.message}`);
    } finally {
      setRelatedLoading(false);
    }
  };

  const formatCompetition = (comp: any): string => {
    if (comp === null || comp === undefined) return "-";
    if (typeof comp === "number") {
      // 숫자를 한글로 변환 (일반적으로 1=낮음, 2=중간, 3=높음)
      if (comp === 1) return "낮음";
      if (comp === 2) return "중간";
      if (comp === 3) return "높음";
      // 범위 기반 변환 (0-33=낮음, 34-66=중간, 67-100=높음)
      if (comp >= 0 && comp <= 33) return "낮음";
      if (comp >= 34 && comp <= 66) return "중간";
      if (comp >= 67 && comp <= 100) return "높음";
      return comp.toString();
    }
    if (typeof comp === "string") {
      const upper = comp.toUpperCase();
      if (upper === "LOW" || comp === "낮음" || comp === "낮은") return "낮음";
      if (upper === "MID" || upper === "MEDIUM" || comp === "중간" || comp === "보통") return "중간";
      if (upper === "HIGH" || comp === "높음" || comp === "높은") return "높음";
      return comp;
    }
    return "-";
  };

  const handleSort = (field: string) => {
    if (sortField === field) {
      setSortOrder(sortOrder === "asc" ? "desc" : "asc");
    } else {
      setSortField(field);
      setSortOrder("desc");
    }
  };

  const getSortedResults = () => {
    if (!sortField) return results;

    const sorted = [...results].sort((a, b) => {
      let aVal: any = a[sortField];
      let bVal: any = b[sortField];

      // null/undefined 처리
      if (aVal == null) aVal = 0;
      if (bVal == null) bVal = 0;

      // 숫자 비교
      if (typeof aVal === "number" && typeof bVal === "number") {
        return sortOrder === "asc" ? aVal - bVal : bVal - aVal;
      }

      // 문자열 비교
      const aStr = String(aVal);
      const bStr = String(bVal);
      if (sortOrder === "asc") {
        return aStr.localeCompare(bStr, "ko");
      } else {
        return bStr.localeCompare(aStr, "ko");
      }
    });

    return sorted;
  };

  const getSortIcon = (field: string) => {
    if (sortField !== field) return "↑↓";
    return sortOrder === "asc" ? "↑" : "↓";
  };

  const showInfo = (title: string, description: string) => {
    setInfoModal({ title, description });
  };

  // 정렬 아이콘 버튼 컴포넌트 (원본의 tt 함수에 해당)
  const SortButton = ({ field }: { field: string }) => (
    <button
      className="sort-btn"
      onClick={() => handleSort(field)}
      aria-label="정렬"
    >
      <span className="sort-icon-text" aria-hidden="true">
        {getSortIcon(field)}
      </span>
    </button>
  );

  const handleExtract = async (e?: React.MouseEvent) => {
    e?.preventDefault();
    e?.stopPropagation();

    // 중복 클릭 방지
    if (extractingRef.current || extracting) {
      return;
    }

    const seeds = seedInput
      .split(",")
      .map((s) => s.trim())
      .filter((s) => s !== "");
    if (seeds.length === 0) {
      alert("시드 키워드를 입력하세요.");
      return;
    }

    extractingRef.current = true;
    setExtracting(true);
    setResults([]);
    setProgress({ current: 0, total: 0, message: "검색 시작 중..." });

    try {
      // 원본처럼 설정에 seeds를 저장
      const config = await api.getConfig();
      const updatedConfig = models.main.Config.createFrom(config);
      updatedConfig.seeds = seeds;
      await api.saveConfig(updatedConfig);

      console.log("검색 시작:", seeds);

      // Promise와 이벤트를 동시에 처리
      const extractPromise = api.runAutoExtract({ seeds });

      // 결과를 기다리는 동안 진행 상황은 이벤트로 받음
      const extractResults = await extractPromise;
      console.log("검색 결과:", extractResults);

      if (!extractResults || extractResults.length === 0) {
        setProgress(null);
        alert("검색 결과가 없습니다. 키워드를 확인해주세요.");
        extractingRef.current = false;
        setExtracting(false);
        return;
      }

      setResults(extractResults);
      // 기본 정렬: 점수 내림차순
      setSortField("score");
      setSortOrder("desc");
      setProgress(null);
      await loadUsage();
    } catch (error: any) {
      console.error("검색 에러:", error);
      setProgress(null);
      const errorMessage =
        error?.message || error?.toString() || "알 수 없는 오류";
      alert(`검색 실패: ${errorMessage}`);
    } finally {
      extractingRef.current = false;
      setExtracting(false);
    }
  };

  return (
    <section className="card">
      <div className="row between">
        <div className="row">
          <h2>
            <span className="title-chip">황금 키워드 검색</span>
          </h2>
          <span className="muted inline-note">
            총 검색량 1천 미만은 제외되며, 각 그룹별 최대 25개까지 조회합니다.
          </span>
        </div>
        <div className="row">
          <button className="btn" onClick={handleReset}>
            검색값 초기화
          </button>
          <div className="usage-pill">
            오늘 사용량 {usage.toLocaleString()} 키워드
          </div>
        </div>
      </div>
      <div className="row seed-row">
        <input
          className="input"
          value={seedInput}
          onChange={(e) => setSeedInput(e.target.value)}
          placeholder="시드 키워드를 입력하세요 (쉼표로 구분)"
          disabled={extracting}
        />
        <button
          className="btn btn-primary"
          onClick={(e) => handleExtract(e)}
          disabled={extracting || !seedInput.trim()}
          type="button"
        >
          검색
        </button>
        {extracting && <span className="status-pill running">진행 중</span>}
        {!extracting && results.length > 0 && (
          <span className="status-pill done">진행 완료</span>
        )}
        {progress && progress.total && (
          <span className="progress-text">
            {progress.current}/{progress.total} -{" "}
            {progress.message || progress.stage}
            {progress.keyword && ` (${progress.keyword})`}
          </span>
        )}
      </div>

      <table className="table">
        <thead>
          <tr className="table-group">
            <th rowSpan={2} className="boxed-col">
              키워드
            </th>
            <th rowSpan={2} className="score-header boxed-col">
              <span className="sort-header">
                <button
                  className="score-info"
                  onClick={() => setShowScoreInfo(true)}
                >
                  <span>키워드<br />점수</span>
                </button>
                <SortButton field="score" />
              </span>
            </th>
            <th colSpan={4} className="search-group-header">
              <span className="badge badge-search">검색 API</span>
            </th>
            <th colSpan={3} className="ad-group-header">
              <span className="badge badge-ad">광고 API</span>
            </th>
            <th rowSpan={2} className="boxed-col">
              <span className="sort-header">
                <button
                  className="category-info"
                  onClick={() => showInfo("경쟁정도", "검색광고 키워드 도구에서 제공하는 경쟁지수입니다. 값의 범위/표현(숫자 또는 LOW/MID/HIGH)은 네이버 API에 따릅니다.")}
                >
                  경쟁<br />정도
                </button>
                <SortButton field="competition" />
              </span>
            </th>
            <th rowSpan={2} className="boxed-col">
              연관
              <br />
              검색어
            </th>
          </tr>
          <tr className="table-sub">
            <th className="search-col search-group-start">
              <span className="sort-header">
                <button
                  className="category-info"
                  onClick={() => showInfo("검색량(PC)", "검색 API 기준 PC 월간 검색량입니다.")}
                >
                  검색량(PC)
                </button>
                <SortButton field="searchPc" />
              </span>
            </th>
            <th className="search-col">
              <span className="sort-header">
                <button
                  className="category-info"
                  onClick={() => showInfo("검색량(모바일)", "검색 API 기준 모바일 월간 검색량입니다.")}
                >
                  검색량<br />(모바일)
                </button>
                <SortButton field="searchMobile" />
              </span>
            </th>
            <th className="search-col">
              <span className="sort-header">
                <button
                  className="category-info"
                  onClick={() => showInfo("총 검색량", "PC/모바일 검색량 합산 월간 검색량입니다.")}
                >
                  총 검색량
                </button>
                <SortButton field="totalSearch" />
              </span>
            </th>
            <th className="search-col search-group-end">
              <span className="sort-header">
                <button
                  className="category-info"
                  onClick={() => showInfo("문서수(블로그)", "네이버 블로그 검색 결과 문서 수입니다.")}
                >
                  문서수<br />(블로그)
                </button>
                <SortButton field="docCount" />
              </span>
            </th>
            <th className="ad-col ad-group-start">
              <span className="sort-header">
                <button
                  className="category-info"
                  onClick={() => showInfo("월간검색수", "검색광고 키워드 도구 기준 월간 검색수입니다.")}
                >
                  월간<br />검색수
                </button>
                <SortButton field="monthlySearch" />
              </span>
            </th>
            <th className="ad-col">
              <span className="sort-header">
                <button
                  className="category-info"
                  onClick={() => showInfo("월평균 클릭수", "검색광고 키워드 도구 기준 월평균 클릭수(PC+모바일 합산)입니다.")}
                >
                  월평균<br />클릭수
                </button>
                <SortButton field="monthlyAvgClicks" />
              </span>
            </th>
            <th className="ad-col ad-group-end">
              <span className="sort-header">
                <button
                  className="category-info"
                  onClick={() => showInfo("월평균 클릭률", "검색광고 키워드 도구 기준 월평균 CTR(%)입니다. PC/모바일 CTR은 클릭수 기준으로 가중 평균됩니다.")}
                >
                  월평균<br />클릭률
                </button>
                <SortButton field="monthlyAvgCtr" />
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          {results.length === 0 && (
            <tr>
              <td colSpan={11} className="empty">
                결과가 없습니다.
              </td>
            </tr>
          )}
          {getSortedResults().map((item, idx) => (
            <tr key={idx}>
              <td>{item.keyword}</td>
              <td>{item.score?.toFixed(2) || "0.00"}</td>
              <td>
                {item.searchPc != null
                  ? item.searchPc.toLocaleString("ko-KR")
                  : "0"}
              </td>
              <td>
                {item.searchMobile != null
                  ? item.searchMobile.toLocaleString("ko-KR")
                  : "0"}
              </td>
              <td>
                {item.totalSearch != null
                  ? item.totalSearch.toLocaleString("ko-KR")
                  : "0"}
              </td>
              <td>
                {item.docCount != null
                  ? item.docCount.toLocaleString("ko-KR")
                  : "0"}
              </td>
              <td>
                {item.monthlySearch != null
                  ? item.monthlySearch.toLocaleString("ko-KR")
                  : "-"}
              </td>
              <td>
                {item.monthlyAvgClicks != null
                  ? item.monthlyAvgClicks.toLocaleString("ko-KR")
                  : "-"}
              </td>
              <td>
                {item.monthlyAvgCtr != null
                  ? `${Number(item.monthlyAvgCtr).toFixed(2)}%`
                  : "-"}
              </td>
              <td>{formatCompetition(item.competition)}</td>
              <td>
                <button
                  className={`btn btn-sm ${
                    relatedStatus[item.keyword] === "loading"
                      ? "btn-warning"
                      : relatedStatus[item.keyword] === "done"
                        ? "btn-success"
                        : "btn-outline"
                  }`}
                  onClick={() => handleAnalyzeRelated(item.keyword)}
                  disabled={relatedStatus[item.keyword] === "loading"}
                >
                  {relatedStatus[item.keyword] === "loading"
                    ? "진행 중"
                    : relatedStatus[item.keyword] === "done"
                      ? "완료"
                      : "보기"}
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {relatedKeyword && (
        <div className="result-card">
          <div className="result-header">
            <div>
              <h3>연관검색어 - {relatedKeyword}</h3>
            </div>
            <button
              className="btn"
              onClick={() => {
                setRelatedKeyword("");
                setRelatedItems([]);
              }}
            >
              닫기
            </button>
          </div>
          {!relatedLoading && relatedItems.length === 0 && (
            <p className="muted">연관검색어가 없습니다.</p>
          )}
          {relatedItems.length > 0 && (
            <table className="mini-table">
              <thead>
                <tr>
                  <th>키워드</th>
                  <th>PC</th>
                  <th>모바일</th>
                  <th>총합</th>
                  <th>문서수</th>
                </tr>
              </thead>
              <tbody>
                {relatedItems.map((item: any, idx: number) => (
                  <tr key={idx}>
                    <td>{item.keyword}</td>
                    <td>{item.searchPc?.toLocaleString() || "0"}</td>
                    <td>{item.searchMobile?.toLocaleString() || "0"}</td>
                    <td>{item.totalSearch?.toLocaleString() || "0"}</td>
                    <td>{item.docCount?.toLocaleString() || "0"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* 정보 모달 */}
      {infoModal && (
        <div className="modal-overlay" onClick={() => setInfoModal(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>{infoModal.title}</h3>
              <button className="modal-close" onClick={() => setInfoModal(null)}>×</button>
            </div>
            <div className="modal-body">
              <p>{infoModal.description}</p>
            </div>
          </div>
        </div>
      )}

      {/* 키워드 점수 정보 모달 */}
      {showScoreInfo && (
        <div className="modal-overlay" onClick={() => setShowScoreInfo(false)}>
          <div className="modal-content score-info-modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>키워드 점수</h3>
              <button className="modal-close" onClick={() => setShowScoreInfo(false)}>×</button>
            </div>
            <div className="modal-body">
              <p>키워드 점수는 검색량, 문서수, 경쟁도 등을 종합하여 산출한 지표입니다.</p>
              <p>점수가 높을수록 블로그 상위 노출에 유리한 키워드입니다.</p>
              <ul>
                <li>검색량이 높고 문서수가 적을수록 점수가 높습니다.</li>
                <li>경쟁도가 낮을수록 점수가 높습니다.</li>
              </ul>
            </div>
          </div>
        </div>
      )}
    </section>
  );
}

// 트렌드 데이터 검색 Page
function TrendDataSearchPage() {
  const [collecting, setCollecting] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [progress, setProgress] = useState<any>(null);
  const [status, setStatus] = useState("");
  const [activeTab, setActiveTab] = useState<"search" | "main">("search");

  useEffect(() => {
    const unsubscribeProgress = api.onCreatorProgress((progress: Progress) => {
      setProgress(progress);
    });
    const unsubscribeDone = api.onCreatorDone(
      (data: { jobId: string; result: any }) => {
        setResult(data.result);
        setCollecting(false);
        setStatus("수집 완료");
      },
    );
    const unsubscribeError = api.onCreatorError(
      (err: { jobId: string; message: string }) => {
        setStatus(`오류: ${err.message}`);
        setCollecting(false);
      },
    );

    return () => {
      unsubscribeProgress();
      unsubscribeDone();
      unsubscribeError();
    };
  }, []);

  const handleCollect = async () => {
    setCollecting(true);
    setResult(null);
    setStatus("");
    setProgress(null);

    try {
      const config = await api.getConfig();
      if (
        !config?.creatorAdvisor?.loginId ||
        !config?.creatorAdvisor?.password
      ) {
        alert("설정에서 네이버 대표 계정을 먼저 입력하세요.");
        setCollecting(false);
        return;
      }

      const response = await api.collectCreatorAdvisor({
        loginId: config.creatorAdvisor.loginId,
        password: config.creatorAdvisor.password,
        showBrowser: false,
      });

      if (!response.ok) {
        setStatus(`오류: ${response.message || "수집 시작 실패"}`);
        setCollecting(false);
      }
    } catch (error: any) {
      setStatus(`오류: ${error.message || "수집 실패"}`);
      setCollecting(false);
    }
  };

  return (
    <section className="card">
      <div className="row between">
        <h2>
          <span className="title-chip">트렌드 데이터 검색</span>
        </h2>
        <span className="muted inline-note">
          실시간 인기있는 주제별, 성별, 연령별 데이터를 제공합니다
        </span>
      </div>
      <div className="row end">
        <button
          className="btn btn-primary"
          onClick={handleCollect}
          disabled={collecting}
        >
          검색
        </button>
        {collecting && <span className="status-pill running">수집 중</span>}
        {collecting && progress && progress.total && (
          <span className="progress-text">
            {progress.current}/{progress.total} - {progress.message}
          </span>
        )}
        {status && <span className="status">{status}</span>}
      </div>
      {collecting && (
        <div className="trend-overlay">
          <div className="app-overlay-card">
            <div className="hourglass" aria-hidden="true"></div>
            <p>약 20초 소요됩니다</p>
          </div>
        </div>
      )}
      {result && (
        <div className="result-card">
          <div className="result-header">
            <h3>수집 결과</h3>
          </div>
          {(result.searchInflowTrends && result.searchInflowTrends.length > 0) ||
          (result.mainInflowContents && result.mainInflowContents.length > 0) ? (
            <div className="result-block">
              <div className="trend-tabs">
                <button
                  className={`tab ${activeTab === "search" ? "active" : ""}`}
                  onClick={() => setActiveTab("search")}
                >
                  검색 유입 트렌드
                </button>
                <button
                  className={`tab ${activeTab === "main" ? "active" : ""}`}
                  onClick={() => setActiveTab("main")}
                >
                  메인 유입 트렌드
                </button>
              </div>
              {activeTab === "search" && (
                <>
                  {result.searchInflowMessage && (
                    <p className="muted">{result.searchInflowMessage}</p>
                  )}
                  {result.searchInflowTrends &&
                    result.searchInflowTrends.length > 0 && (
                      <>
                        {/* 주제별 섹션 */}
                        {result.searchInflowTrends.some((trend: any) => {
                          const normalizeSource = (source: string) =>
                            String(source || "").replace(/\s/g, "");
                          return normalizeSource(trend.source) === "주제별인기유입검색어";
                        }) && (
                          <div className="trend-section-group">
                            <h4 className="trend-section-title">주제별</h4>
                            <div className="trend-cards">
                              {result.searchInflowTrends
                                .filter((trend: any) => {
                                  const normalizeSource = (source: string) =>
                                    String(source || "").replace(/\s/g, "");
                                  return normalizeSource(trend.source) === "주제별인기유입검색어";
                                })
                                .map((trend: any, idx: number) => (
                                  <div key={idx} className="trend-section topic">
                                    <div className="trend-card">
                                      <h4>{trend.category}</h4>
                                      <ul className="trend-list">
                                        {trend.items?.map((item: any, itemIdx: number) => (
                                          <li key={itemIdx}>
                                            <span className="trend-keyword">
                                              {item.keyword}
                                            </span>
                                            <span
                                              className={`trend-change ${item.status === "up" ? "up" : item.status === "down" ? "down" : "new"}`}
                                            >
                                              {item.change}
                                            </span>
                                          </li>
                                        ))}
                                      </ul>
                                    </div>
                                  </div>
                                ))}
                            </div>
                          </div>
                        )}
                        {/* 성별, 연령별 섹션 */}
                        {result.searchInflowTrends.some((trend: any) => {
                          const normalizeSource = (source: string) =>
                            String(source || "").replace(/\s/g, "");
                          return normalizeSource(trend.source) === "성별,연령별인기유입검색어";
                        }) && (
                          <div className="trend-section-group">
                            <h4 className="trend-section-title">성별, 연령별</h4>
                            <div className="trend-cards">
                              {result.searchInflowTrends
                                .filter((trend: any) => {
                                  const normalizeSource = (source: string) =>
                                    String(source || "").replace(/\s/g, "");
                                  return normalizeSource(trend.source) === "성별,연령별인기유입검색어";
                                })
                                .slice(0, 4)
                                .map((trend: any, idx: number) => (
                                  <div key={idx} className="trend-section demographic">
                                    <div className="trend-card">
                                      <h4>{trend.category}</h4>
                                      <ul className="trend-list">
                                        {trend.items?.map((item: any, itemIdx: number) => (
                                          <li key={itemIdx}>
                                            <span className="trend-keyword">
                                              {item.keyword}
                                            </span>
                                            <span
                                              className={`trend-change ${item.status === "up" ? "up" : item.status === "down" ? "down" : "new"}`}
                                            >
                                              {item.change}
                                            </span>
                                          </li>
                                        ))}
                                      </ul>
                                    </div>
                                  </div>
                                ))}
                            </div>
                          </div>
                        )}
                      </>
                    )}
                  {result.searchInflowTrends &&
                    result.searchInflowTrends.length === 0 &&
                    !result.searchInflowMessage && (
                      <p className="muted">검색 유입 트렌드 데이터가 없습니다.</p>
                    )}
                </>
              )}
              {activeTab === "main" && (
                <>
                  {result.mainInflowContents &&
                    result.mainInflowContents.length > 0 && (
                      <table className="mini-table">
                        <thead>
                          <tr>
                            <th>순위</th>
                            <th>콘텐츠 제목</th>
                          </tr>
                        </thead>
                        <tbody>
                          {result.mainInflowContents.map(
                            (item: any, idx: number) => (
                              <tr key={idx}>
                                <td>{item.rank || idx + 1}</td>
                                <td>
                                  {item.url ? (
                                    <a
                                      href={item.url}
                                      target="_blank"
                                      rel="noopener noreferrer"
                                      className="link-inline"
                                    >
                                      {item.title}
                                    </a>
                                  ) : (
                                    item.title
                                  )}
                                </td>
                              </tr>
                            ),
                          )}
                        </tbody>
                      </table>
                    )}
                  {(!result.mainInflowContents ||
                    result.mainInflowContents.length === 0) && (
                    <p className="muted">메인 유입 트렌드 데이터가 없습니다.</p>
                  )}
                </>
              )}
            </div>
          ) : (
            <p className="muted">수집된 데이터가 없습니다.</p>
          )}
        </div>
      )}
    </section>
  );
}

// 단일 키워드 검색 Page
function SingleKeywordSearchPage() {
  const [keyword, setKeyword] = useState("");
  const [analyzing, setAnalyzing] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [progress, setProgress] = useState<any>(null);
  const [status, setStatus] = useState("");
  const [infoModal, setInfoModal] = useState<{ title: string; description: string } | null>(null);

  const showInfo = (title: string, description: string) => {
    setInfoModal({ title, description });
  };

  useEffect(() => {
    const unsubscribe = api.onKeywordProgress((progress: Progress) => {
      setProgress(progress);
    });

    const handleSingleKeywordSearch = (e: any) => {
      setKeyword(e.detail);
      handleAnalyze(e.detail);
    };

    window.addEventListener(
      "singleKeywordSearch",
      handleSingleKeywordSearch as EventListener,
    );

    return () => {
      unsubscribe();
      window.removeEventListener(
        "singleKeywordSearch",
        handleSingleKeywordSearch as EventListener,
      );
    };
  }, []);

  const formatCompetition = (comp: any): string => {
    if (comp === null || comp === undefined) return "-";
    if (typeof comp === "number") {
      // 숫자를 한글로 변환 (일반적으로 1=낮음, 2=중간, 3=높음)
      if (comp === 1) return "낮음";
      if (comp === 2) return "중간";
      if (comp === 3) return "높음";
      // 범위 기반 변환 (0-33=낮음, 34-66=중간, 67-100=높음)
      if (comp >= 0 && comp <= 33) return "낮음";
      if (comp >= 34 && comp <= 66) return "중간";
      if (comp >= 67 && comp <= 100) return "높음";
      return comp.toString();
    }
    if (typeof comp === "string") {
      const upper = comp.toUpperCase();
      if (upper === "LOW" || comp === "낮음" || comp === "낮은") return "낮음";
      if (upper === "MID" || upper === "MEDIUM" || comp === "중간" || comp === "보통") return "중간";
      if (upper === "HIGH" || comp === "높음" || comp === "높은") return "높음";
      return comp;
    }
    return "-";
  };

  const handleAnalyze = async (kw?: string) => {
    const targetKeyword = kw || keyword;
    if (!targetKeyword.trim()) {
      alert("키워드를 입력하세요.");
      return;
    }

    setAnalyzing(true);
    setResult(null);
    setStatus("");
    setProgress(null);

    try {
      const analysisResult = await api.analyzeKeyword(targetKeyword);
      setResult(analysisResult);
      setStatus("검색 완료");
    } catch (error: any) {
      setStatus(`오류: ${error.message}`);
      alert(`검색 실패: ${error.message}`);
    } finally {
      setAnalyzing(false);
    }
  };

  return (
    <section className="card search">
      <div className="row between">
        <h2>
          <span className="title-chip">단일 키워드 검색</span>
        </h2>
        <span className="muted inline-note">
          1개의 키워드에 대한 자세한 검색값을 제공합니다
        </span>
      </div>
      <div className="row">
        <input
          className="input"
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          placeholder="키워드를 입력하세요"
          disabled={analyzing}
        />
        <button
          className="btn btn-primary"
          onClick={() => handleAnalyze()}
          disabled={analyzing || !keyword.trim()}
        >
          검색
        </button>
        {analyzing && <span className="status-pill running">수집 중</span>}
        {analyzing && progress && progress.total && (
          <span className="progress-text">
            {progress.current}/{progress.total} -{" "}
            {progress.message || progress.stage}
          </span>
        )}
        {status && !analyzing && <span className="status">{status}</span>}
      </div>
      {result && (
        <div className="result-card">
          <div className="result-header">
            <h3>{result.keyword}</h3>
            <div className="score-badge">
              {result.score?.toFixed(2) || "0.00"}
            </div>
          </div>
          <div className="result-block">
            <h4 className="section-gap">
              <span className="section-label badge-search">검색 API</span>
            </h4>
            <table className="mini-table">
              <thead>
                <tr>
                  <th>그룹</th>
                  <th>
                    <button className="category-info" onClick={() => showInfo("검색량(PC)", "검색 API 기준 PC 월간 검색량입니다.")}>
                      검색량(PC)
                    </button>
                  </th>
                  <th>
                    <button className="category-info" onClick={() => showInfo("검색량(모바일)", "검색 API 기준 모바일 월간 검색량입니다.")}>
                      검색량(모바일)
                    </button>
                  </th>
                  <th>
                    <button className="category-info" onClick={() => showInfo("총 검색량", "PC/모바일 검색량 합산 월간 검색량입니다.")}>
                      총 검색량
                    </button>
                  </th>
                  <th>
                    <button className="category-info" onClick={() => showInfo("문서수(블로그)", "네이버 블로그 검색 결과 문서 수입니다.")}>
                      문서수(블로그)
                    </button>
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>{result.group || "-"}</td>
                  <td>{result.searchPc != null ? result.searchPc.toLocaleString("ko-KR") : "0"}</td>
                  <td>{result.searchMobile != null ? result.searchMobile.toLocaleString("ko-KR") : "0"}</td>
                  <td>{result.totalSearch != null ? result.totalSearch.toLocaleString("ko-KR") : "0"}</td>
                  <td>{result.docCount != null ? result.docCount.toLocaleString("ko-KR") : "0"}</td>
                </tr>
              </tbody>
            </table>
            <h4 className="section-gap">
              <span className="section-label badge-ad">광고 API</span>
            </h4>
            <table className="mini-table">
              <thead>
                <tr>
                  <th>
                    <button className="category-info" onClick={() => showInfo("월간검색수", "검색광고 키워드 도구 기준 월간 검색수입니다.")}>
                      월간검색수
                    </button>
                  </th>
                  <th>
                    <button className="category-info" onClick={() => showInfo("월평균 클릭수", "검색광고 키워드 도구 기준 월평균 클릭수(PC+모바일 합산)입니다.")}>
                      월평균 클릭수
                    </button>
                  </th>
                  <th>
                    <button className="category-info" onClick={() => showInfo("월평균 클릭률", "검색광고 키워드 도구 기준 월평균 CTR(%)입니다. PC/모바일 CTR은 클릭수 기준으로 가중 평균됩니다.")}>
                      월평균 클릭률
                    </button>
                  </th>
                  <th>
                    <button className="category-info" onClick={() => showInfo("경쟁정도", "검색광고 키워드 도구에서 제공하는 경쟁지수입니다. 값의 범위/표현(숫자 또는 LOW/MID/HIGH)은 네이버 API에 따릅니다.")}>
                      경쟁정도
                    </button>
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>{result.monthlySearch != null ? result.monthlySearch.toLocaleString("ko-KR") : "-"}</td>
                  <td>{result.monthlyAvgClicks != null ? result.monthlyAvgClicks.toLocaleString("ko-KR") : "-"}</td>
                  <td>{result.monthlyAvgCtr != null ? `${Number(result.monthlyAvgCtr).toFixed(2)}%` : "-"}</td>
                  <td>{formatCompetition(result.competition)}</td>
                </tr>
              </tbody>
            </table>
            {result.relatedDetails && result.relatedDetails.length > 0 && (
              <>
                <h4 className="section-gap">연관검색어</h4>
                <table className="mini-table">
                  <thead>
                    <tr>
                      <th>키워드</th>
                      <th>PC</th>
                      <th>모바일</th>
                      <th>총합</th>
                      <th>문서수</th>
                    </tr>
                  </thead>
                  <tbody>
                    {result.relatedDetails.map((item: any, idx: number) => (
                      <tr key={idx}>
                        <td>{item.keyword}</td>
                        <td>
                          {item.searchPc != null
                            ? item.searchPc.toLocaleString("ko-KR")
                            : "0"}
                        </td>
                        <td>
                          {item.searchMobile != null
                            ? item.searchMobile.toLocaleString("ko-KR")
                            : "0"}
                        </td>
                        <td>
                          {item.totalSearch != null
                            ? item.totalSearch.toLocaleString("ko-KR")
                            : "0"}
                        </td>
                        <td>
                          {item.docCount != null
                            ? item.docCount.toLocaleString("ko-KR")
                            : "0"}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </>
            )}
          </div>
        </div>
      )}

      {/* 정보 모달 */}
      {infoModal && (
        <div className="modal-overlay" onClick={() => setInfoModal(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>{infoModal.title}</h3>
              <button className="modal-close" onClick={() => setInfoModal(null)}>×</button>
            </div>
            <div className="modal-body">
              <p>{infoModal.description}</p>
            </div>
          </div>
        </div>
      )}
    </section>
  );
}

// Creator Page (not used in main UI, but kept for reference)
function CreatorPage() {
  const [loginId, setLoginId] = useState("");
  const [password, setPassword] = useState("");
  const [showBrowser, setShowBrowser] = useState(false);
  const [collecting, setCollecting] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [progress, setProgress] = useState<any>(null);
  const [error, setError] = useState<string>("");

  useEffect(() => {
    const unsubscribeProgress = api.onCreatorProgress((progress: Progress) => {
      setProgress(progress);
    });
    const unsubscribeDone = api.onCreatorDone(
      (data: { jobId: string; result: any }) => {
        setResult(data.result);
        setCollecting(false);
      },
    );
    const unsubscribeError = api.onCreatorError(
      (err: { jobId: string; message: string }) => {
        setError(err.message);
        setCollecting(false);
      },
    );

    return () => {
      unsubscribeProgress();
      unsubscribeDone();
      unsubscribeError();
    };
  }, []);

  const handleCollect = async () => {
    if (!loginId.trim() || !password.trim()) {
      alert("네이버 ID와 비밀번호를 입력하세요.");
      return;
    }

    setCollecting(true);
    setResult(null);
    setError("");
    setProgress(null);

    try {
      const response = await api.collectCreatorAdvisor({
        loginId,
        password,
        showBrowser,
      });

      if (!response.ok) {
        setError(response.message || "수집 시작 실패");
        setCollecting(false);
      }
      // 성공 시 이벤트로 결과를 받음
    } catch (error: any) {
      setError(error.message || "수집 실패");
      setCollecting(false);
    }
  };

  return (
    <div className="creator-page">
      <h2>크리에이터 어드바이저 수집</h2>
      <div className="input-section">
        <div className="form-group">
          <label>네이버 ID</label>
          <input
            type="text"
            value={loginId}
            onChange={(e) => setLoginId(e.target.value)}
            placeholder="네이버 ID"
            disabled={collecting}
          />
        </div>
        <div className="form-group">
          <label>비밀번호</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="비밀번호"
            disabled={collecting}
          />
        </div>
        <div className="form-group">
          <label>
            <input
              type="checkbox"
              checked={showBrowser}
              onChange={(e) => setShowBrowser(e.target.checked)}
              disabled={collecting}
            />
            브라우저 표시 (디버깅용)
          </label>
        </div>
        <button
          onClick={handleCollect}
          disabled={collecting}
          className="collect-button"
        >
          {collecting ? "수집 중..." : "수집 시작"}
        </button>
      </div>

      {error && <div className="error-message">{error}</div>}

      {progress && (
        <div className="progress-section">
          <div className="progress-bar">
            <div
              className="progress-fill"
              style={{ width: `${(progress.current / progress.total) * 100}%` }}
            />
          </div>
          <div className="progress-text">
            {progress.current}/{progress.total} - {progress.message}
            {progress.detail && ` (${progress.detail})`}
          </div>
        </div>
      )}

      {result && (
        <div className="result-section">
          <h3>수집 결과</h3>
          <div className="result-card">
            <div className="result-row">
              <strong>URL:</strong>{" "}
              <a href={result.url} target="_blank" rel="noreferrer">
                {result.url}
              </a>
            </div>
            {result.searchInflowMessage && (
              <div className="result-row">
                <strong>메시지:</strong> {result.searchInflowMessage}
              </div>
            )}
            {result.searchInflowTrends &&
              result.searchInflowTrends.length > 0 && (
                <div className="trends-section">
                  <h4>검색 유입 트렌드</h4>
                  {result.searchInflowTrends.map((trend: any, idx: number) => (
                    <div key={idx} className="trend-group">
                      <h5>
                        {trend.category} ({trend.source})
                      </h5>
                      <table className="result-table">
                        <thead>
                          <tr>
                            <th>키워드</th>
                            <th>변동</th>
                            <th>상태</th>
                          </tr>
                        </thead>
                        <tbody>
                          {trend.items.map((item: any, itemIdx: number) => (
                            <tr key={itemIdx}>
                              <td>{item.keyword}</td>
                              <td>{item.change}</td>
                              <td>{item.status}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  ))}
                </div>
              )}
            {result.mainInflowContents &&
              result.mainInflowContents.length > 0 && (
                <div className="main-inflow-section">
                  <h4>메인 유입 콘텐츠</h4>
                  {result.mainInflowContents.map(
                    (content: any, idx: number) => (
                      <div key={idx} className="content-item">
                        <strong>{content.title}</strong>
                        <p>{content.text}</p>
                      </div>
                    ),
                  )}
                </div>
              )}
          </div>
        </div>
      )}
    </div>
  );
}

export default App;
