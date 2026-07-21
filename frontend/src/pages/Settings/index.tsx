import React from 'react';

const Settings: React.FC = () => {
  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold text-text-primary mb-6">Настройки</h1>
      
      <div className="card mb-6">
        <h2 className="text-xl font-semibold text-text-primary mb-4">
          Управление фондами
        </h2>
        <p className="text-text-secondary">В разработке...</p>
      </div>

      <div className="card mb-6">
        <h2 className="text-xl font-semibold text-text-primary mb-4">
          Настройки LLM и поиска
        </h2>
        <p className="text-text-secondary">В разработке...</p>
      </div>

      <div className="card">
        <h2 className="text-xl font-semibold text-text-primary mb-4">
          Данные
        </h2>
        <p className="text-text-secondary">В разработке...</p>
      </div>
    </div>
  );
};

export default Settings;
