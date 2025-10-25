import React, { useState, useMemo } from 'react'
import { Line } from 'react-chartjs-2'
import type { Expense, Income, ChartPeriod, ChartType } from '../../types'

type ExpenseLineChartProps = {
  expenses: Expense[]
  incomes?: Income[]
  chartType?: ChartType
  chartPeriod?: ChartPeriod
  onChartTypeChange?: (type: ChartType) => void
  onChartPeriodChange?: (period: ChartPeriod) => void
}

export const ExpenseLineChart: React.FC<ExpenseLineChartProps> = ({ 
  expenses, 
  incomes = [], 
  chartType = 'expenses',
  chartPeriod = 'days',
  onChartTypeChange,
  onChartPeriodChange
}) => {
  const [localChartType, setLocalChartType] = useState<ChartType>(chartType)
  const [localChartPeriod, setLocalChartPeriod] = useState<ChartPeriod>(chartPeriod)

  const currentChartType = chartType || localChartType
  const currentChartPeriod = chartPeriod || localChartPeriod

  const prepareLineChartData = useMemo(() => {
    const getDateRange = () => {
      const now = new Date()
      const dataPoints = 14 // Максимум 14 значений
      
      switch (currentChartPeriod) {
        case 'days':
          return [...Array(dataPoints)].map((_, i) => {
            const d = new Date(now)
            d.setDate(d.getDate() - (dataPoints - 1 - i))
            return d.toISOString().split('T')[0]
          })
        case 'months':
          return [...Array(dataPoints)].map((_, i) => {
            const d = new Date(now)
            d.setMonth(d.getMonth() - (dataPoints - 1 - i))
            return d.toISOString().substring(0, 7) // YYYY-MM
          })
        case 'years':
          return [...Array(Math.min(dataPoints, 5))].map((_, i) => {
            const d = new Date(now)
            d.setFullYear(d.getFullYear() - (Math.min(dataPoints, 5) - 1 - i))
            return d.getFullYear().toString()
          })
        default:
          return []
      }
    }

    const dateRange = getDateRange()
    
    const getDataForPeriod = (data: (Expense | Income)[], period: string) => {
      return data
        .filter((item) => {
          const itemDate = item.timestamp.split('T')[0]
          if (currentChartPeriod === 'days') {
            return itemDate === period
          } else if (currentChartPeriod === 'months') {
            return itemDate.startsWith(period)
          } else if (currentChartPeriod === 'years') {
            return itemDate.startsWith(period)
          }
          return false
        })
        .reduce((sum, item) => sum + item.amount_cents, 0) / 100
    }

    const formatLabel = (period: string) => {
      if (currentChartPeriod === 'days') {
        return new Date(period).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' })
      } else if (currentChartPeriod === 'months') {
        return new Date(period + '-01').toLocaleDateString('ru-RU', { month: 'short', year: '2-digit' })
      } else if (currentChartPeriod === 'years') {
        return period
      }
      return period
    }

    const datasets = []

    if (currentChartType === 'expenses' || currentChartType === 'both') {
      const expensesData = dateRange.map(period => getDataForPeriod(expenses, period))
      datasets.push({
        label: 'Расходы (₽)',
        data: expensesData,
        borderColor: 'rgb(239, 68, 68)',
        backgroundColor: 'rgba(239, 68, 68, 0.1)',
        tension: 0.3,
        fill: false
      })
    }

    if (currentChartType === 'incomes' || currentChartType === 'both') {
      const incomesData = dateRange.map(period => getDataForPeriod(incomes, period))
      datasets.push({
        label: 'Приходы (₽)',
        data: incomesData,
        borderColor: 'rgb(16, 185, 129)',
        backgroundColor: 'rgba(16, 185, 129, 0.1)',
        tension: 0.3,
        fill: false
      })
    }

    return {
      labels: dateRange.map(formatLabel),
      datasets
    }
  }, [expenses, incomes, currentChartType, currentChartPeriod])

  const handleChartTypeChange = (type: ChartType) => {
    setLocalChartType(type)
    onChartTypeChange?.(type)
  }

  const handleChartPeriodChange = (period: ChartPeriod) => {
    setLocalChartPeriod(period)
    onChartPeriodChange?.(period)
  }

  const getChartTitle = () => {
    const periodText = currentChartPeriod === 'days' ? 'дни' : 
                      currentChartPeriod === 'months' ? 'месяцы' : 'годы'
    const typeText = currentChartType === 'expenses' ? 'Расходы' :
                    currentChartType === 'incomes' ? 'Приходы' : 'Расходы и приходы'
    return `${typeText} по ${periodText}`
  }

  return (
    <div className="glass-card chart-container">
      <div className="chart-header">
        <h3>{getChartTitle()}</h3>
        <div className="chart-controls">
          <div className="chart-type-buttons">
            <button
              className={currentChartType === 'expenses' ? 'active' : 'secondary'}
              onClick={() => handleChartTypeChange('expenses')}
            >
              Расходы
            </button>
            <button
              className={currentChartType === 'incomes' ? 'active' : 'secondary'}
              onClick={() => handleChartTypeChange('incomes')}
            >
              Приходы
            </button>
            <button
              className={currentChartType === 'both' ? 'active' : 'secondary'}
              onClick={() => handleChartTypeChange('both')}
            >
              Оба
            </button>
          </div>
          <div className="chart-period-buttons">
            <button
              className={currentChartPeriod === 'days' ? 'active' : 'secondary'}
              onClick={() => handleChartPeriodChange('days')}
            >
              Дни
            </button>
            <button
              className={currentChartPeriod === 'months' ? 'active' : 'secondary'}
              onClick={() => handleChartPeriodChange('months')}
            >
              Месяцы
            </button>
            <button
              className={currentChartPeriod === 'years' ? 'active' : 'secondary'}
              onClick={() => handleChartPeriodChange('years')}
            >
              Годы
            </button>
          </div>
        </div>
      </div>
      <Line
        data={prepareLineChartData}
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

