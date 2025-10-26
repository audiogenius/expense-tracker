import { useState, useEffect } from 'react'
import { fetchFamilyGroups } from '../../api'

type FamilySettingsProps = {
  token: string
}

type GroupMember = {
  user_id: number
  username: string
}

type TelegramGroup = {
  group_id: number
  group_name: string
  chat_type: string
  members: GroupMember[]
}

type FamilyGroupsResponse = {
  groups: TelegramGroup[]
}

export const FamilySettings: React.FC<FamilySettingsProps> = ({ token }) => {
  const [groups, setGroups] = useState<TelegramGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadGroups()
  }, [])

  const loadGroups = async () => {
    try {
      setLoading(true)
      setError(null)
      const response: FamilyGroupsResponse = await fetchFamilyGroups(token)
      setGroups(response.groups || [])
    } catch (err) {
      setError('Не удалось загрузить данные о семье')
      console.error('Failed to load groups:', err)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="family-settings">
        <div className="loading-indicator">
          <div className="spinner"></div>
          <span>Загрузка...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="family-settings">
        <div className="error-message">
          ⚠️ {error}
        </div>
      </div>
    )
  }

  return (
    <div className="family-settings">
      <div className="family-info">
        <h2>👨‍👩‍👧 Управление семьёй</h2>
        <p className="subtitle">
          Здесь вы можете управлять участниками семейной группы и настроить доступ к расходам
        </p>
      </div>

      {groups.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state-icon">👥</div>
          <h3>Нет семейных групп</h3>
          <p>
            Семейные группы создаются автоматически, когда вы добавляете расход в группе Telegram.
          </p>
          <div className="info-box">
            <h4>💡 Как это работает:</h4>
            <ul>
              <li>Добавьте бота в группу Telegram с вашей семьёй</li>
              <li>Запишите расход в группе через бота</li>
              <li>Группа автоматически появится здесь</li>
              <li>Вы сможете управлять участниками и их доступом</li>
            </ul>
          </div>
        </div>
      ) : (
        <div className="groups-list">
          {groups.map((group) => (
            <div key={group.group_id} className="group-card">
              <div className="group-header">
                <h3>{group.group_name}</h3>
                <span className="group-type">{group.chat_type}</span>
              </div>
              
              <div className="group-members">
                <h4>Участники ({group.members.length})</h4>
                <div className="members-list">
                  {group.members.map((member) => (
                    <div key={member.user_id} className="member-item">
                      <div className="member-info">
                        <span className="member-name">@{member.username}</span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      <div className="family-features">
        <h3>🔜 Скоро появится:</h3>
        <ul>
          <li>✨ Добавление участников вручную</li>
          <li>🔒 Настройка прав доступа</li>
          <li>👁️ Управление видимостью расходов</li>
          <li>📊 Статистика по участникам</li>
        </ul>
      </div>
    </div>
  )
}

