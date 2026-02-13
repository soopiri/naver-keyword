# 키워드 황금 추출기 (Wails 버전)

Electron 앱을 Wails로 포팅한 키워드 분석 도구입니다.

## 구현된 기능

✅ 설정 관리 (GetConfig, SaveConfig)
✅ 사용량 추적 (GetUsage)
✅ 키워드 분석 (AnalyzeKeyword)
✅ 자동 키워드 추출 (RunAutoExtract)
✅ 네이버 SearchAd API 연동
✅ 네이버 OpenAPI 연동
✅ 스코어링 로직
✅ 이벤트 시스템 (진행 상황 전달)
✅ 로그 관리

⚠️ Playwright 자동화 기능은 아직 구현되지 않았습니다 (creator.go 참조)

## 설치 및 실행

### 1. 의존성 설치

```bash
# Go 모듈 다운로드
go mod download

# Wails 개발 도구 설치 (필요시)
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 2. 개발 모드 실행

```bash
# 프론트엔드와 백엔드를 함께 실행
wails dev
```

### 3. 빌드

```bash
# 프로덕션 빌드
wails build
```

## Playwright 자동화 기능

Playwright 자동화 기능이 구현되어 있습니다! 사용하려면 브라우저를 설치해야 합니다.

### 브라우저 설치

```bash
# playwright-go가 이미 설치되어 있으므로 브라우저만 설치하면 됩니다
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install chromium
```

또는 모든 브라우저 설치:

```bash
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install
```

### 구현된 기능

✅ 네이버 로그인 자동화
✅ 블로그 홈/내 블로그 접속
✅ 통계 페이지 접속
✅ 크리에이터 어드바이저 접속
✅ 검색 유입 트렌드 수집 (주제별, 성별/연령별)
✅ 메인 유입 콘텐츠 수집
✅ 슬라이드 조작 및 데이터 추출

## 설정 파일 위치

설정 파일은 다음 위치에 저장됩니다:
- macOS/Linux: `~/.pick-keyword/config.json`
- Windows: `%USERPROFILE%\.pick-keyword\config.json`

로그 파일은 `~/.pick-keyword/logs/` 디렉토리에 저장됩니다.

## API 사용법

프론트엔드에서 `frontend/src/api.ts`를 import하여 사용할 수 있습니다:

```typescript
import { api } from './api';

// 설정 가져오기
const config = await api.getConfig();

// 키워드 분석
const result = await api.analyzeKeyword('키워드');

// 진행 상황 리스너
const unsubscribe = api.onKeywordProgress((progress) => {
  console.log(`진행: ${progress.current}/${progress.total}`);
});
```

## 주의사항

1. **Playwright 기능**: 현재 Playwright 자동화 기능은 구현되지 않았습니다. `creator.go`를 구현해야 합니다.

2. **남은시간 계산**: 요청하신 대로 남은시간 계산 기능은 제외되었습니다.

3. **타입 생성**: Wails는 `wails dev` 또는 `wails build` 실행 시 자동으로 TypeScript 타입을 생성합니다. `frontend/wailsjs/go/main/App.d.ts` 파일이 자동 생성됩니다.

4. **이벤트 리스너**: 이벤트 리스너는 컴포넌트 언마운트 시 정리해야 합니다:

```typescript
useEffect(() => {
  const unsubscribe = api.onKeywordProgress(handleProgress);
  return () => unsubscribe();
}, []);
```

## 문제 해결

### 타입 에러가 발생하는 경우

`wails dev`를 실행하면 자동으로 타입이 생성됩니다. 타입 파일이 없으면:

```bash
wails dev
```

### Go 컴파일 에러

필요한 패키지가 설치되지 않았을 수 있습니다:

```bash
go mod tidy
go mod download
```

## 라이센스

원본 Electron 앱의 라이센스를 따릅니다.
