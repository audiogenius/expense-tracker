import React from 'react'
import type { Balance, Period } from '../../types'

type BalanceCardProps = {
  balance: Balance | null
  filterPeriod: Period
  onPeriodChange: (period: Period) => void
}

export const BalanceCard: React.FC<BalanceCardProps> = ({ balance, filterPeriod, onPeriodChange }) => {
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
      </div>
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

