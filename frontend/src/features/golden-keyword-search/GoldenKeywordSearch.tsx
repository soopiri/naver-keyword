// GoldenKeywordSearch Feature - 완결된 UI 제공
import { useState, useEffect, useRef } from "react";
import { keywordApi } from "../../entities/keyword";
import { configApi } from "../../entities/config";
import { wailsEvents } from "../../shared/api/wails";
import { formatCompetition } from "../../shared/utils/format";
import type { Progress, AutoExtractResult, RelatedDetail } from "../../entities/keyword/types";
import * as models from "../../../wailsjs/go/models";

export function GoldenKeywordSearch() {
  const [seedInput, setSeedInput] = useState("");
  const [extracting, setExtracting] = useState(false);
  const [results, setResults] = useState<AutoExtractResult[]>([]);
  const [progress, setProgress] = useState<Progress | null>(null);
  const [usage, setUsage] = useState(0);
  const [relatedKeyword, setRelatedKeyword] = useState<string>("");
  const [relatedLoading, setRelatedLoading] = useState(false);
  const [relatedItems, setRelatedItems] = useState<RelatedDetail[]>([]);
  const [relatedStatus, setRelatedStatus] = useState<Record<string, string>>({});
  const extractingRef = useRef(false);
  const [sortField, setSortField] = useState<string | null>(null);
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");
  const [infoModal, setInfoModal] = useState<{ title: string; description: string } | null>(null);
  const [showScoreInfo, setShowScoreInfo] = useState(false);

  useEffect(() => {
    const unsubscribe = wailsEvents.onAutoProgress((progress: Progress) => {
      setProgress(progress);
    });
    return () => unsubscribe();
  }, []);

  useEffect(() => {
    loadUsage();
  }, []);

  const loadUsage = async () => {
    try {
      const usageData = await keywordApi.getUsage();
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
      const result = await keywordApi.analyze(keyword);
      setRelatedItems(result.relatedDetails || []);
      setRelatedStatus((prev) => ({ ...prev, [keyword]: "done" }));
    } catch (error: any) {
      setRelatedStatus((prev) => ({ ...prev, [keyword]: "error" }));
      alert(`연관검색어 수집 실패: ${error.message}`);
    } finally {
      setRelatedLoading(false);
    }
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
      let aVal: any = (a as any)[sortField];
      let bVal: any = (b as any)[sortField];

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
      const config = await configApi.get();
      const updatedConfig = models.main.Config.createFrom(config);
      updatedConfig.seeds = seeds;
      await configApi.save(updatedConfig);

      console.log("검색 시작:", seeds);

      const extractPromise = keywordApi.runAutoExtract({ seeds });
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
              <td>{(item as any).keyword}</td>
              <td>{((item as any).score?.toFixed(2)) || "0.00"}</td>
              <td>
                {(item as any).searchPc != null
                  ? (item as any).searchPc.toLocaleString("ko-KR")
                  : "0"}
              </td>
              <td>
                {(item as any).searchMobile != null
                  ? (item as any).searchMobile.toLocaleString("ko-KR")
                  : "0"}
              </td>
              <td>
                {(item as any).totalSearch != null
                  ? (item as any).totalSearch.toLocaleString("ko-KR")
                  : "0"}
              </td>
              <td>
                {(item as any).docCount != null
                  ? (item as any).docCount.toLocaleString("ko-KR")
                  : "0"}
              </td>
              <td>
                {(item as any).monthlySearch != null
                  ? (item as any).monthlySearch.toLocaleString("ko-KR")
                  : "-"}
              </td>
              <td>
                {(item as any).monthlyAvgClicks != null
                  ? (item as any).monthlyAvgClicks.toLocaleString("ko-KR")
                  : "-"}
              </td>
              <td>
                {(item as any).monthlyAvgCtr != null
                  ? `${Number((item as any).monthlyAvgCtr).toFixed(2)}%`
                  : "-"}
              </td>
              <td>{formatCompetition((item as any).competition)}</td>
              <td>
                <button
                  className={`btn btn-sm ${
                    relatedStatus[(item as any).keyword] === "loading"
                      ? "btn-warning"
                      : relatedStatus[(item as any).keyword] === "done"
                        ? "btn-success"
                        : "btn-outline"
                  }`}
                  onClick={() => handleAnalyzeRelated((item as any).keyword)}
                  disabled={relatedStatus[(item as any).keyword] === "loading"}
                >
                  {relatedStatus[(item as any).keyword] === "loading"
                    ? "진행 중"
                    : relatedStatus[(item as any).keyword] === "done"
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
