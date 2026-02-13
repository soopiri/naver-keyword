// SingleKeywordSearch Feature - 완결된 UI 제공
import { useState, useEffect } from "react";
import { keywordApi } from "../../entities/keyword";
import { wailsEvents } from "../../shared/api/wails";
import { formatCompetition } from "../../shared/utils/format";
import type { Progress, KeywordAnalysisResult } from "../../entities/keyword/types";

export function SingleKeywordSearch() {
  const [keyword, setKeyword] = useState("");
  const [analyzing, setAnalyzing] = useState(false);
  const [result, setResult] = useState<KeywordAnalysisResult | null>(null);
  const [progress, setProgress] = useState<Progress | null>(null);
  const [status, setStatus] = useState("");
  const [infoModal, setInfoModal] = useState<{ title: string; description: string } | null>(null);

  const showInfo = (title: string, description: string) => {
    setInfoModal({ title, description });
  };

  useEffect(() => {
    const unsubscribe = wailsEvents.onKeywordProgress((progress: Progress) => {
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
      const analysisResult = await keywordApi.analyze(targetKeyword);
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
