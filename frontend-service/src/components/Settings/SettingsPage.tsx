import { useState } from 'react'
import { CategoriesPage } from '../Categories/CategoriesPage'
import { FamilySettings } from './FamilySettings'
import { TransactionManagement } from './TransactionManagement'

type SettingsPageProps = {
  token: string
  onBack: () => void
}

type SettingsTab = 'family' | 'categories' | 'transactions'

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
        </div>
      </div>

      <div className="settings-content">
        {activeTab === 'family' && <FamilySettings token={token} />}
        {activeTab === 'categories' && <CategoriesPage token={token} />}
        {activeTab === 'transactions' && <TransactionManagement token={token} />}
      </div>
    </div>
  )
}

