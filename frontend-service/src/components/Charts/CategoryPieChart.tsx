import React from 'react'
import { Pie } from 'react-chartjs-2'
import type { Expense, Category } from '../../types'
import { getCategoryName } from '../../utils/helpers'

type CategoryPieChartProps = {
  expenses: Expense[]
  categories: Category[]
}

export const CategoryPieChart: React.FC<CategoryPieChartProps> = ({ expenses, categories }) => {
  const preparePieChartData = () => {
    const expensesByCategory: Record<string, number> = {}

    expenses.forEach((e) => {
      const catName = getCategoryName(e.category_id, categories)
      expensesByCategory[catName] = (expensesByCategory[catName] || 0) + e.amount_cents / 100
    })

    const colors = [
      'rgba(124, 58, 237, 0.8)',
      'rgba(59, 130, 246, 0.8)',
      'rgba(16, 185, 129, 0.8)',
      'rgba(245, 158, 11, 0.8)',
      'rgba(239, 68, 68, 0.8)',
      'rgba(168, 85, 247, 0.8)',
      'rgba(6, 182, 212, 0.8)',
      'rgba(251, 146, 60, 0.8)'
    ]

    return {
      labels: Object.keys(expensesByCategory),
      datasets: [
        {
          data: Object.values(expensesByCategory),
          backgroundColor: colors
        }
      ]
    }
  }

  return (
    <div className="glass-card chart-container">
      <h3>Категории расходов</h3>
      <Pie
        data={preparePieChartData()}
        options={{
          responsive: true,
          maintainAspectRatio: true,
          plugins: {
            legend: {
              labels: {
                color: '#ffffff',
                font: { size: 14, weight: 'bold' }
              }
            }
          }
        }}
      />
    </div>
  )
}

