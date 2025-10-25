import React, { useState } from 'react'
import type { Balance, Period, CustomPeriod } from '../../types'

type BalanceCardProps = {
  balance: Balance | null
  filterPeriod: Period
  customPeriod?: CustomPeriod
  onPeriodChange: (period: Period) => void
  onCustomPeriodChange?: (period: CustomPeriod) => void
}

export const BalanceCard: React.FC<BalanceCardProps> = ({ 
  balance, 
  filterPeriod, 
  customPeriod,
  onPeriodChange, 
  onCustomPeriodChange 
}) => {
  const [showCustomPeriod, setShowCustomPeriod] = useState(false)
  const [tempCustomPeriod, setTempCustomPeriod] = useState<CustomPeriod>({
    start_date: customPeriod?.start_date || '',
    end_date: customPeriod?.end_date || ''
  })

  const handleCustomPeriodSubmit = () => {
    if (tempCustomPeriod.start_date && tempCustomPeriod.end_date) {
      onCustomPeriodChange?.(tempCustomPeriod)
      onPeriodChange('custom')
      setShowCustomPeriod(false)
    }
  }

  const handleCustomPeriodCancel = () => {
    setTempCustomPeriod({
      start_date: customPeriod?.start_date || '',
      end_date: customPeriod?.end_date || ''
    })
    setShowCustomPeriod(false)
  }

  return (
    <div className="glass-card">
      <h3>Семейный бюджет</h3>
      <div className="period-buttons" style={{ marginBottom: 20 }}>
        <button
          className={filterPeriod === 'all' ? 'active' : 'secondary'}
          onClick={() => onPeriodChange('all')}
        >
          Все
        </button>
        <button
          className={filterPeriod === 'day' ? 'active' : 'secondary'}
          onClick={() => onPeriodChange('day')}
        >
          День
        </button>
        <button
          className={filterPeriod === 'week' ? 'active' : 'secondary'}
          onClick={() => onPeriodChange('week')}
        >
          Неделя
        </button>
        <button
          className={filterPeriod === 'month' ? 'active' : 'secondary'}
          onClick={() => onPeriodChange('month')}
        >
          Месяц
        </button>
        <button
          className={filterPeriod === 'custom' ? 'active' : 'secondary'}
          onClick={() => setShowCustomPeriod(!showCustomPeriod)}
        >
          Выбрать период
        </button>
      </div>

      {showCustomPeriod && (
        <div className="custom-period-selector">
          <div className="date-inputs">
            <div className="date-input-group">
              <label>От:</label>
              <input
                type="date"
                value={tempCustomPeriod.start_date}
                onChange={(e) => setTempCustomPeriod(prev => ({ ...prev, start_date: e.target.value }))}
                className="date-input"
              />
            </div>
            <div className="date-input-group">
              <label>До:</label>
              <input
                type="date"
                value={tempCustomPeriod.end_date}
                onChange={(e) => setTempCustomPeriod(prev => ({ ...prev, end_date: e.target.value }))}
                className="date-input"
              />
            </div>
          </div>
          <div className="custom-period-buttons">
            <button onClick={handleCustomPeriodSubmit} className="btn-primary">
              Применить
            </button>
            <button onClick={handleCustomPeriodCancel} className="btn-secondary">
              Отмена
            </button>
          </div>
        </div>
      )}
      {balance && (
        <div className="balance-grid">
          <div className="balance-card">
            <div className="balance-label">Баланс</div>
            <div
              className="balance-value"
              style={{ color: balance.balance_rubles >= 0 ? '#10b981' : '#ef4444' }}
            >
              {balance.balance_rubles.toFixed(2)} ₽
            </div>
          </div>
          <div className="balance-stats">
            <div className="balance-stat">
              <span className="balance-stat-label">Доходы:</span>
              <span className="balance-stat-value" style={{ color: '#10b981' }}>
                +{balance.total_incomes_rubles.toFixed(2)} ₽
              </span>
            </div>
            <div className="balance-stat">
              <span className="balance-stat-label">Расходы:</span>
              <span className="balance-stat-value" style={{ color: '#ef4444' }}>
                -{balance.total_expenses_rubles.toFixed(2)} ₽
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

