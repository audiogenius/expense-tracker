import React from 'react'
import type { Income } from '../../types'
import { getIncomeTypeName, formatDate, formatCurrency } from '../../utils/helpers'

type IncomesListProps = {
  incomes: Income[]
}

export const IncomesList: React.FC<IncomesListProps> = ({ incomes }) => {
  return (
    <div className="glass-card">
      <h3>Последние приходы</h3>
      <ul className="expenses">
        {incomes.length === 0 ? (
          <li style={{ textAlign: 'center', color: 'var(--text-muted)', padding: '40px 20px' }}>
            Нет приходов
          </li>
        ) : (
          incomes.slice(0, 20).map((inc) => (
            <li key={inc.id} className="expense-item">
              <div className="expense-info">
                <div className="expense-date">{formatDate(inc.timestamp)}</div>
                <div className="expense-category">{getIncomeTypeName(inc.income_type)}</div>
                {inc.description && (
                  <div
                    className="income-description"
                    style={{ fontSize: '13px', color: 'var(--text-muted)', marginTop: 4 }}
                  >
                    {inc.description}
                  </div>
                )}
                {inc.username && (
                  <div
                    className="income-user"
                    style={{ fontSize: '12px', color: 'var(--accent)', marginTop: 2 }}
                  >
                    @{inc.username}
                  </div>
                )}
              </div>
              <div className="expense-amount" style={{ color: '#10b981' }}>
                +{formatCurrency(inc.amount_cents)}
              </div>
            </li>
          ))
        )}
      </ul>
    </div>
  )
}

