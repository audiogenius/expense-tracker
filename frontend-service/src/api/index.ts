import axios from 'axios'
import type { Expense, Income, Category, Balance } from '../types'

const API_BASE = '/api'

// Auth
export const loginWithTelegram = async (authData: Record<string, any>) => {
  const res = await axios.post(`${API_BASE}/login`, authData)
  return {
    token: res.data.token,
    profile: {
      username: res.data.username,
      id: res.data.id,
      photo_url: res.data.photo_url
    }
  }
}

// Categories
export const fetchCategories = async (): Promise<Category[]> => {
  const res = await axios.get(`${API_BASE}/categories`)
  return res.data || []
}

// Expenses
export const fetchExpenses = async (token: string): Promise<Expense[]> => {
  const res = await axios.get(`${API_BASE}/expenses`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return res.data || []
}

export const fetchTotalExpenses = async (token: string, period: string) => {
  const res = await axios.get(`${API_BASE}/expenses/total?period=${period}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return res.data
}

export const addExpense = async (token: string, amountCents: number, categoryId: number | null) => {
  await axios.post(`${API_BASE}/expenses`, {
    amount_cents: amountCents,
    category_id: categoryId,
    timestamp: new Date().toISOString()
  }, {
    headers: { Authorization: `Bearer ${token}` }
  })
}

// Incomes
export const fetchIncomes = async (token: string): Promise<Income[]> => {
  const res = await axios.get(`${API_BASE}/incomes`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return res.data || []
}

export const addIncome = async (
  token: string,
  amountCents: number,
  incomeType: string,
  description: string
) => {
  await axios.post(`${API_BASE}/incomes`, {
    amount_cents: amountCents,
    income_type: incomeType,
    description,
    timestamp: new Date().toISOString()
  }, {
    headers: { Authorization: `Bearer ${token}` }
  })
}

// Balance
export const fetchBalance = async (token: string, period: string): Promise<Balance> => {
  const res = await axios.get(`${API_BASE}/balance?period=${period}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return res.data
}

