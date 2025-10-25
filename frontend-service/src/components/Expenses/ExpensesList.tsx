import React from 'react'
import type { Expense, Category } from '../../types'
import { getCategoryName, formatDate, formatCurrency } from '../../utils/helpers'

type ExpensesListProps = {
  expenses: Expense[]
  categories: Category[]
  filterCategory: number | null
  onFilterChange: (categoryId: number | null) => void
}

export const ExpensesList: React.FC<ExpensesListProps> = ({
  expenses,
  categories,
  filterCategory,
  onFilterChange
}) => {
  return (
    <div className="glass-card">
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 20,
          flexWrap: 'wrap',
          gap: '12px'
        }}
      >
        <h3 style={{ margin: 0 }}>Последние расходы</h3>
        <select
          value={filterCategory || ''}
          onChange={(e) => onFilterChange(e.target.value ? parseInt(e.target.value) : null)}
        >
          <option value="">Все категории</option>
          {categories.map((cat) => (
            <option key={cat.id} value={cat.id}>
              {cat.name}
            </option>
          ))}
        </select>
      </div>
      <ul className="expenses">
        {expenses.length === 0 ? (
          <li style={{ textAlign: 'center', color: 'var(--text-muted)', padding: '40px 20px' }}>
            Нет расходов
          </li>
        ) : (
          expenses.map((e) => (
            <li key={e.id} className="expense-item">
              <div className="expense-info">
                <div className="expense-date">{formatDate(e.timestamp)}</div>
                <div className="expense-category">{getCategoryName(e.category_id, categories)}</div>
                {e.username && (
                  <div
                    className="expense-user"
                    style={{ fontSize: '12px', color: 'var(--accent)', marginTop: 2 }}
                  >
                    @{e.username}
                  </div>
                )}
              </div>
              <div className="expense-amount">{formatCurrency(e.amount_cents)}</div>
            </li>
          ))
        )}
      </ul>
    </div>
  )
}

