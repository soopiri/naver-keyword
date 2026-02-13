// KeywordSearchSection Widget - features 조합만 (로컬 상태 금지)
import { SingleKeywordSearch } from "../../features/single-keyword-search/SingleKeywordSearch";
import { GoldenKeywordSearch } from "../../features/golden-keyword-search/GoldenKeywordSearch";
import { TrendDataSearch } from "../../features/trend-data-search/TrendDataSearch";

export function KeywordSearchSection() {
  return (
    <>
      <SingleKeywordSearch />
      <GoldenKeywordSearch />
      <TrendDataSearch />
    </>
  );
}
