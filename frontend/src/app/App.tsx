// App Shell - 라우터, providers, layout
import { useState, useEffect } from "react";
import "../style.css";
import { configApi } from "../entities/config";
import { Settings } from "../features/settings/Settings";
import { MainPage } from "../pages/MainPage";
import type { Config } from "../entities/config";

function App() {
  const [showSettings, setShowSettings] = useState(false);
  const [config, setConfig] = useState<Config | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      const loadedConfig = await configApi.get();
      setConfig(loadedConfig);
    } catch (error) {
      console.error("설정 로드 실패:", error);
    } finally {
      setLoading(false);
    }
  };

  const saveConfig = async (newConfig: Config) => {
    try {
      await configApi.save(newConfig);
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
        <Settings config={config} onSave={saveConfig} />
      )}

      <MainPage />
    </div>
  );
}

export default App;
