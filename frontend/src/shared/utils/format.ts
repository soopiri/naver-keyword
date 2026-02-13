// 공용 포맷 유틸리티

export const formatCompetition = (comp: any): string => {
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
