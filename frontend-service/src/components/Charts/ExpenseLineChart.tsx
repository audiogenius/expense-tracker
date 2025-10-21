import React from 'react'
import { Line } from 'react-chartjs-2'
import type { Expense } from '../../types'

type ExpenseLineChartProps = {
  expenses: Expense[]
}

export const ExpenseLineChart: React.FC<ExpenseLineChartProps> = ({ expenses }) => {
  const prepareLineChartData = () => {
    const last7Days = [...Array(7)].map((_, i) => {
      const d = new Date()
      d.setDate(d.getDate() - (6 - i))
      return d.toISOString().split('T')[0]
    })

    const expensesByDay = last7Days.map((day) => {
      return expenses
        .filter((e) => e.timestamp.split('T')[0] === day)
        .reduce((sum, e) => sum + e.amount_cents, 0) / 100
    })

    return {
      labels: last7Days.map((d) =>
        new Date(d).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' })
      ),
      datasets: [
        {
          label: 'Расходы (₽)',
          data: expensesByDay,
          borderColor: 'rgb(124, 58, 237)',
          backgroundColor: 'rgba(124, 58, 237, 0.1)',
          tension: 0.3
        }
      ]
    }
  }

  return (
    <div className="glass-card chart-container">
      <h3>Расходы за последние 7 дней</h3>
      <Line
        data={prepareLineChartData()}
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
          },
          scales: {
            x: {
              ticks: { color: '#ffffff' },
              grid: { color: 'rgba(255, 255, 255, 0.1)' }
            },
            y: {
              ticks: { color: '#ffffff' },
              grid: { color: 'rgba(255, 255, 255, 0.1)' }
            }
          }
        }}
      />
    </div>
  )
}

