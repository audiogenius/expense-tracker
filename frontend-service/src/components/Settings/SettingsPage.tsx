import { useState } from 'react'
import { CategoriesPage } from '../Categories/CategoriesPage'
import { FamilySettings } from './FamilySettings'
import { TransactionManagement } from './TransactionManagement'
import { AIAssistant } from './AIAssistant'

type SettingsPageProps = {
  token: string
  onBack: () => void
}

type SettingsTab = 'family' | 'categories' | 'transactions' | 'ai'

export const SettingsPage: React.FC<SettingsPageProps> = ({ token, onBack }) => {
  const [activeTab, setActiveTab] = useState<SettingsTab>('family')

  return (
    <div className="settings-page">
      <div className="settings-header">
        <div className="settings-title-row">
          <button onClick={onBack} className="back-btn">
            ← Назад
          </button>
          <h1>⚙️ Настройки</h1>
        </div>
        
        <div className="settings-tabs">
          <button
            className={`settings-tab ${activeTab === 'family' ? 'active' : ''}`}
            onClick={() => setActiveTab('family')}
          >
            👨‍👩‍👧 Семья
          </button>
          <button
            className={`settings-tab ${activeTab === 'categories' ? 'active' : ''}`}
            onClick={() => setActiveTab('categories')}
          >
            📁 Категории
          </button>
          <button
            className={`settings-tab ${activeTab === 'transactions' ? 'active' : ''}`}
            onClick={() => setActiveTab('transactions')}
          >
            🗑️ Управление операциями
          </button>
          <button
            className={`settings-tab ${activeTab === 'ai' ? 'active' : ''}`}
            onClick={() => setActiveTab('ai')}
          >
            🤖 AI Помощник
          </button>
        </div>
      </div>

      <div className="settings-content">
        {activeTab === 'family' && <FamilySettings token={token} />}
        {activeTab === 'categories' && <CategoriesPage token={token} editable={true} />}
        {activeTab === 'transactions' && <TransactionManagement token={token} />}
        {activeTab === 'ai' && <AIAssistant token={token} />}
      </div>
    </div>
  )
}

