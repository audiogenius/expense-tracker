import React, { useEffect, useState } from 'react'
import './styles.css'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
} from 'chart.js'

// Types
import type { Expense, Income, Category, Balance, Profile, Period, CustomPeriod, ChartType, ChartPeriod } from './types'

// API
import * as api from './api'

// Components
import { TelegramLogin } from './components/Auth/TelegramLogin'
import { Header } from './components/Header/Header'
import { BalanceCard } from './components/Balance/BalanceCard'
import { ExpenseLineChart } from './components/Charts/ExpenseLineChart'
import { CategoryPieChart } from './components/Charts/CategoryPieChart'
import { ExpensesList } from './components/Expenses/ExpensesList'
import { AddExpense } from './components/Expenses/AddExpense'
import { IncomesList } from './components/Incomes/IncomesList'
import { AddIncome } from './components/Incomes/AddIncome'
import { RecentTransactions } from './components/Transactions/RecentTransactions'
import { AddTransaction } from './components/Transactions/AddTransaction'
import { TransactionsPage } from './components/Transactions/TransactionsPage'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, ArcElement)

const App: React.FC = () => {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'))
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [incomes, setIncomes] = useState<Income[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [balance, setBalance] = useState<Balance | null>(null)
  const [profile, setProfile] = useState<Profile | null>(() => {
    try {
      return JSON.parse(localStorage.getItem('profile') || 'null')
    } catch {
      return null
    }
  })

  // Filters
  const [filterCategory, setFilterCategory] = useState<number | null>(null)
  const [filterPeriod, setFilterPeriod] = useState<Period>('all')
  const [customPeriod, setCustomPeriod] = useState<CustomPeriod | undefined>(undefined)
  const [currentPage, setCurrentPage] = useState<'dashboard' | 'transactions'>('dashboard')
  
  // Chart settings
  const [chartType, setChartType] = useState<ChartType>('expenses')
  const [chartPeriod, setChartPeriod] = useState<ChartPeriod>('days')

  // Fetch data on mount and period change
  useEffect(() => {
    if (token) {
      fetchData()
    }
  }, [token, filterPeriod])

  const fetchData = async () => {
    if (!token) return
    try {
      const [expensesData, incomesData, categoriesData, balanceData] = await Promise.all([
        api.fetchExpenses(token),
        api.fetchIncomes(token),
        api.fetchCategories(),
        api.fetchBalance(token, filterPeriod)
      ])
      setExpenses(expensesData)
      setIncomes(incomesData)
      setCategories(categoriesData)
      setBalance(balanceData)
    } catch (err) {
      console.error('fetchData error', err)
    }
  }

  const handleTelegramAuth = async (authData: Record<string, any>) => {
    try {
      const { token: newToken, profile: newProfile } = await api.loginWithTelegram(authData)
      localStorage.setItem('token', newToken)
      localStorage.setItem('profile', JSON.stringify(newProfile))
      setProfile(newProfile)
      setToken(newToken)
    } catch (err) {
      alert('Login failed: ' + String(err))
    }
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('profile')
    setToken(null)
    setExpenses([])
    setIncomes([])
    setProfile(null)
    setBalance(null)
  }

  const handleAddExpense = async (amount: string, categoryId: number | null) => {
    if (!token) {
      alert('not authenticated')
      return
    }
    const cents = Math.round(parseFloat(amount || '0') * 100)
    if (cents <= 0) {
      alert('Amount must be positive')
      return
    }
    try {
      await api.addExpense(token, cents, categoryId)
      await fetchData()
    } catch (err) {
      console.error('addExpense error', err)
      alert('Add expense failed')
    }
  }

  const handleAddIncome = async (amount: string, type: string, description: string) => {
    if (!token) {
      alert('not authenticated')
      return
    }
    const cents = Math.round(parseFloat(amount || '0') * 100)
    if (cents <= 0) {
      alert('Сумма должна быть положительной')
      return
    }
    try {
      await api.addIncome(token, cents, type, description)
      await fetchData()
    } catch (err) {
      console.error('addIncome error', err)
      alert('Ошибка добавления прихода')
    }
  }

  const handleTransactionAdded = async () => {
    await fetchData()
  }

  const handleViewAllTransactions = () => {
    setCurrentPage('transactions')
  }

  const handleBackToDashboard = () => {
    setCurrentPage('dashboard')
  }

  // Filter expenses
  const filteredExpenses = expenses.filter((e) => {
    if (filterCategory !== null && e.category_id !== filterCategory) return false

    if (filterPeriod !== 'all') {
      const expenseDate = new Date(e.timestamp)
      const now = new Date()
      const daysDiff = (now.getTime() - expenseDate.getTime()) / (1000 * 60 * 60 * 24)

      if (filterPeriod === 'week' && daysDiff > 7) return false
      if (filterPeriod === 'month' && daysDiff > 30) return false
    }

    return true
  })

  if (!token) {
    return (
      <div className="app-center">
        <TelegramLogin onAuth={handleTelegramAuth} />
      </div>
    )
  }

  if (currentPage === 'transactions') {
    return (
      <div className="app-center">
        <Header profile={profile} onLogout={handleLogout} />
        <div className="page-header">
          <button onClick={handleBackToDashboard} className="back-btn">
            ← Назад к дашборду
          </button>
        </div>
        <TransactionsPage token={token!} />
      </div>
    )
  }

  return (
    <div className="app-center">
      <Header profile={profile} onLogout={handleLogout} />

      <div className="main-grid">
        {/* LEFT COLUMN */}
        <div className="left-column">
          <BalanceCard 
            balance={balance} 
            filterPeriod={filterPeriod} 
            customPeriod={customPeriod}
            onPeriodChange={setFilterPeriod}
            onCustomPeriodChange={setCustomPeriod}
          />

          {/* Charts */}
          {filteredExpenses.length > 0 && (
            <div className="charts-grid">
              <ExpenseLineChart 
                expenses={filteredExpenses} 
                incomes={incomes}
                chartType={chartType}
                chartPeriod={chartPeriod}
                onChartTypeChange={setChartType}
                onChartPeriodChange={setChartPeriod}
              />
              <CategoryPieChart expenses={filteredExpenses} categories={categories} />
            </div>
          )}
        </div>

        {/* RIGHT COLUMN */}
        <div className="right-column">
          <RecentTransactions token={token!} onViewAll={handleViewAllTransactions} />
          <AddTransaction token={token!} onTransactionAdded={handleTransactionAdded} />
          <ExpensesList
            expenses={filteredExpenses}
            categories={categories}
            filterCategory={filterCategory}
            onFilterChange={setFilterCategory}
          />
          <AddExpense categories={categories} onAdd={handleAddExpense} />
          <AddIncome onAdd={handleAddIncome} />
          <IncomesList incomes={incomes} />
        </div>
      </div>
    </div>
  )
}

export default App
