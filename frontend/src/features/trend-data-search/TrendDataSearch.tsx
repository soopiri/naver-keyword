// TrendDataSearch Feature - 완결된 UI 제공
import { useState, useEffect } from "react";
import { keywordApi } from "../../entities/keyword";
import { configApi } from "../../entities/config";
import { wailsEvents } from "../../shared/api/wails";
import type { Progress, CreatorAdvisorResponse } from "../../entities/keyword/types";

export function TrendDataSearch() {
  const [collecting, setCollecting] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [progress, setProgress] = useState<Progress | null>(null);
  const [status, setStatus] = useState("");
  const [activeTab, setActiveTab] = useState<"search" | "main">("search");

  useEffect(() => {
    const unsubscribeProgress = wailsEvents.onCreatorProgress((progress: Progress) => {
      setProgress(progress);
    });
    const unsubscribeDone = wailsEvents.onCreatorDone(
      (data: { jobId: string; result: any }) => {
        setResult(data.result);
        setCollecting(false);
        setStatus("수집 완료");
      },
    );
    const unsubscribeError = wailsEvents.onCreatorError(
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
      const config = await configApi.get();
      if (
        !config?.creatorAdvisor?.loginId ||
        !config?.creatorAdvisor?.password
      ) {
        alert("설정에서 네이버 대표 계정을 먼저 입력하세요.");
        setCollecting(false);
        return;
      }

      const response = await keywordApi.collectCreatorAdvisor({
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

  const normalizeSource = (source: string) =>
    String(source || "").replace(/\s/g, "");

  const getAgeSortValue = (category: string): number => {
    const rangeMatch = category.match(/(\d+)-(\d+)세/);
    if (rangeMatch) {
      const startAge = parseInt(rangeMatch[1], 10);
      const endAge = parseInt(rangeMatch[2], 10);
      return startAge * 1000 + endAge;
    }
    
    const singleMatch = category.match(/(\d+)세-/);
    if (singleMatch) {
      const startAge = parseInt(singleMatch[1], 10);
      return startAge * 1000 + 999;
    }
    
    return 999999;
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
                        {result.searchInflowTrends.some((trend: any) =>
                          normalizeSource(trend.source) === "주제별인기유입검색어"
                        ) && (
                          <div className="trend-section-group">
                            <h4 className="trend-section-title">주제별</h4>
                            <div className="trend-cards">
                              {result.searchInflowTrends
                                .filter((trend: any) =>
                                  normalizeSource(trend.source) === "주제별인기유입검색어"
                                )
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
                        {result.searchInflowTrends.some((trend: any) =>
                          normalizeSource(trend.source) === "성별,연령별인기유입검색어"
                        ) && (() => {
                          const demographicTrends = result.searchInflowTrends.filter((trend: any) =>
                            normalizeSource(trend.source) === "성별,연령별인기유입검색어"
                          );

                          const maleTrends = demographicTrends
                            .filter((trend: any) => trend.category.includes("남자"))
                            .sort((a: any, b: any) => getAgeSortValue(a.category) - getAgeSortValue(b.category));

                          const femaleTrends = demographicTrends
                            .filter((trend: any) => trend.category.includes("여자"))
                            .sort((a: any, b: any) => getAgeSortValue(a.category) - getAgeSortValue(b.category));

                          return (
                            <div className="trend-section-group">
                              <h4 className="trend-section-title">성별, 연령별</h4>
                              
                              {femaleTrends.length > 0 && (
                                <div className="demographic-gender-section">
                                  <h5 className="demographic-gender-title">여자</h5>
                                  <div className="trend-cards">
                                    {femaleTrends.map((trend: any, idx: number) => (
                                      <div key={`female-${idx}`} className="trend-section demographic">
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

                              {maleTrends.length > 0 && (
                                <div className="demographic-gender-section">
                                  <h5 className="demographic-gender-title">남자</h5>
                                  <div className="trend-cards">
                                    {maleTrends.map((trend: any, idx: number) => (
                                      <div key={`male-${idx}`} className="trend-section demographic">
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
                            </div>
                          );
                        })()}
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
