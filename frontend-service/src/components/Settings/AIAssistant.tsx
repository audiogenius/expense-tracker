import { useState } from 'react'

type AIAssistantProps = {
  token: string
}

export const AIAssistant: React.FC<AIAssistantProps> = ({ token }) => {
  const [loading, setLoading] = useState(false)
  const [summary, setSummary] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const generateSummary = async () => {
    try {
      setLoading(true)
      setError(null)
      setSummary(null)

      const response = await fetch('/api/analytics/summary', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          period: 'day'
        })
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }

      const data = await response.json()
      setSummary(data.summary || data.message || 'Саммари получен')
    } catch (err) {
      console.error('Failed to generate summary:', err)
      setError('Не удалось получить саммари. Проверьте, что Ollama запущена и доступна.')
    } finally {
      setLoading(false)
    }
  }

  const testOllama = async () => {
    try {
      setLoading(true)
      setError(null)
      setSummary(null)

      const response = await fetch('/api/analytics/health', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }

      const data = await response.json()
      
      if (data.ollama_status === 'available') {
        setSummary('✅ Ollama доступна и работает!\n\nМодель: ' + (data.model || 'llama2'))
      } else {
        setError('❌ Ollama недоступна. Проверьте, что сервис запущен.')
      }
    } catch (err) {
      console.error('Failed to test Ollama:', err)
      setError('❌ Не удалось подключиться к Ollama. Проверьте, что analytics-service запущен.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="ai-assistant">
      <div className="ai-header">
        <h2>🤖 AI Помощник</h2>
        <p className="subtitle">
          Используйте искусственный интеллект для анализа ваших расходов и получения рекомендаций
        </p>
      </div>

      <div className="ai-actions">
        <button
          onClick={testOllama}
          disabled={loading}
          className="btn btn-test"
        >
          {loading ? '⏳ Проверка...' : '🔍 Проверить Ollama'}
        </button>
        
        <button
          onClick={generateSummary}
          disabled={loading}
          className="btn btn-generate"
        >
          {loading ? '⏳ Генерация...' : '📊 Получить саммари за сегодня'}
        </button>
      </div>

      {error && (
        <div className="ai-error">
          <div className="error-icon">⚠️</div>
          <div className="error-text">{error}</div>
        </div>
      )}

      {summary && (
        <div className="ai-summary">
          <div className="summary-header">
            <h3>📋 Результат</h3>
          </div>
          <div className="summary-content">
            {summary.split('\n').map((line, index) => (
              <p key={index}>{line}</p>
            ))}
          </div>
        </div>
      )}

      <div className="ai-info">
        <h3>💡 Как это работает:</h3>
        <ul>
          <li>
            <strong>Ollama</strong> - это локальный AI сервис, который работает на вашем сервере
          </li>
          <li>
            Нажмите <strong>"Проверить Ollama"</strong>, чтобы убедиться, что сервис доступен
          </li>
          <li>
            Нажмите <strong>"Получить саммари"</strong>, чтобы получить анализ расходов за сегодня
          </li>
          <li>
            AI проанализирует ваши расходы и приходы, выявит паттерны и даст рекомендации
          </li>
        </ul>
      </div>

      <div className="ai-features">
        <h3>🔜 Скоро появится:</h3>
        <ul>
          <li>📊 Саммари за неделю/месяц</li>
          <li>💡 Персональные рекомендации по экономии</li>
          <li>📈 Прогнозы расходов на основе истории</li>
          <li>🎯 Советы по оптимизации бюджета</li>
          <li>📱 Команда в Telegram боте для получения саммари</li>
        </ul>
      </div>
    </div>
  )
}

