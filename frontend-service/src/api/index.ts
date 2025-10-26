import axios from 'axios'
import { apiCache, CACHE_TTL } from '../utils/apiCache'
import type { 
  Expense, 
  Income, 
  Category, 
  Balance, 
  Subcategory, 
  Transaction, 
  TransactionResponse, 
  TransactionFilters, 
  CategorySuggestion 
} from '../types'

const API_BASE = '/api'

// Auth
export const loginWithTelegram = async (authData: Record<string, any>) => {
  // Convert authData to URL-encoded format
  const formData = new URLSearchParams()
  Object.entries(authData).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      formData.append(key, String(value))
    }
  })
  
  const res = await axios.post(`${API_BASE}/login`, formData, {
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    }
  })
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
  const url = `${API_BASE}/categories`
  
  // Check cache first
  const cached = apiCache.get(url)
  if (cached) {
    return cached
  }
  
  const res = await axios.get(url)
  const data = res.data || []
  
  // Cache the result
  apiCache.set(url, data, CACHE_TTL.CATEGORIES)
  
  return data
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

export const addExpense = async (
  token: string, 
  amountCents: number, 
  categoryId: number | null, 
  subcategoryId?: number | null,
  operationType: 'expense' | 'income' = 'expense'
) => {
  await axios.post(`${API_BASE}/expenses`, {
    amount_cents: amountCents,
    category_id: categoryId,
    subcategory_id: subcategoryId,
    operation_type: operationType,
    timestamp: new Date().toISOString()
  }, {
    headers: { Authorization: `Bearer ${token}` }
  })
  
  // Clear cache after adding transaction
  apiCache.clearPattern('/api/transactions')
  apiCache.clearPattern('/api/expenses')
  apiCache.clearPattern('/api/incomes')
  apiCache.clearPattern('/api/balance')
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
export const fetchBalance = async (
  token: string, 
  period: string, 
  customPeriod?: { start_date: string; end_date: string }
): Promise<Balance> => {
  let url = `${API_BASE}/balance?period=${period}`
  if (customPeriod) {
    url += `&start_date=${customPeriod.start_date}&end_date=${customPeriod.end_date}`
  }
  const res = await axios.get(url, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return res.data
}

// Subcategories
export const fetchSubcategories = async (token: string, categoryId?: number): Promise<Subcategory[]> => {
  const url = categoryId 
    ? `${API_BASE}/subcategories?category_id=${categoryId}`
    : `${API_BASE}/subcategories`
  const res = await axios.get(url, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return res.data || []
}

export const createSubcategory = async (
  token: string, 
  name: string, 
  categoryId: number, 
  aliases: string[] = []
) => {
  const res = await axios.post(`${API_BASE}/subcategories`, {
    name,
    category_id: categoryId,
    aliases
  }, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return res.data
}

export const updateSubcategory = async (
  token: string,
  id: number,
  name: string,
  categoryId: number,
  aliases: string[] = []
) => {
  const res = await axios.put(`${API_BASE}/subcategories/${id}`, {
    name,
    category_id: categoryId,
    aliases
  }, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return res.data
}

export const deleteSubcategory = async (token: string, id: number) => {
  await axios.delete(`${API_BASE}/subcategories/${id}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
}

// Transactions
export const fetchTransactions = async (
  token: string, 
  filters: TransactionFilters = {}
): Promise<TransactionResponse> => {
  const params = new URLSearchParams()
  
  if (filters.operation_type) params.append('operation_type', filters.operation_type)
  if (filters.category_id) params.append('category_id', filters.category_id.toString())
  if (filters.subcategory_id) params.append('subcategory_id', filters.subcategory_id.toString())
  if (filters.start_date) params.append('start_date', filters.start_date)
  if (filters.end_date) params.append('end_date', filters.end_date)
  if (filters.cursor) params.append('cursor', filters.cursor)
  if (filters.limit) params.append('limit', filters.limit.toString())

  const url = `${API_BASE}/transactions?${params.toString()}`
  
  // Check cache first (only for non-cursor requests)
  if (!filters.cursor) {
    const cached = apiCache.get(url, { token })
    if (cached) {
      return cached
    }
  }

  const res = await axios.get(url, {
    headers: { Authorization: `Bearer ${token}` }
  })
  
  const data = res.data
  
  // Cache the result (only for non-cursor requests)
  if (!filters.cursor) {
    apiCache.set(url, data, CACHE_TTL.TRANSACTIONS, { token })
  }
  
  return data
}

// Category Suggestions
export const fetchCategorySuggestions = async (
  token: string, 
  query: string
): Promise<CategorySuggestion[]> => {
  const res = await axios.get(`${API_BASE}/suggestions/categories?query=${encodeURIComponent(query)}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return res.data || []
}

// Family Groups
export const fetchFamilyGroups = async (token: string) => {
  const res = await axios.get(`${API_BASE}/family/groups`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return res.data
}

// Soft Delete Transactions
export const softDeleteTransaction = async (token: string, transactionId: number) => {
  await axios.delete(`${API_BASE}/transactions/${transactionId}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  // Clear cache after deletion
  apiCache.clearPattern('/api/transactions')
  apiCache.clearPattern('/api/expenses')
  apiCache.clearPattern('/api/incomes')
  apiCache.clearPattern('/api/balance')
}

export const restoreTransaction = async (token: string, transactionId: number) => {
  await axios.post(`${API_BASE}/transactions/${transactionId}/restore`, {}, {
    headers: { Authorization: `Bearer ${token}` }
  })
  // Clear cache after restoration
  apiCache.clearPattern('/api/transactions')
  apiCache.clearPattern('/api/expenses')
  apiCache.clearPattern('/api/incomes')
  apiCache.clearPattern('/api/balance')
}

export const fetchDeletedTransactions = async (token: string, limit: number = 50) => {
  const res = await axios.get(`${API_BASE}/transactions/deleted?limit=${limit}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return res.data
}

// Category management functions
export const createCategory = async (token: string, name: string, aliases: string[] = []) => {
  const res = await axios.post(`${API_BASE}/categories`, { name, aliases }, {
    headers: { Authorization: `Bearer ${token}` }
  })
  // Clear categories cache
  apiCache.clearPattern('/api/categories')
  return res.data
}

export const updateCategory = async (token: string, id: number, name: string, aliases: string[] = []) => {
  const res = await axios.put(`${API_BASE}/categories/${id}`, { name, aliases }, {
    headers: { Authorization: `Bearer ${token}` }
  })
  // Clear categories cache
  apiCache.clearPattern('/api/categories')
  return res.data
}

export const deleteCategory = async (token: string, id: number) => {
  await axios.delete(`${API_BASE}/categories/${id}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  // Clear categories cache
  apiCache.clearPattern('/api/categories')
}


// Transaction management functions
export const createTransaction = async (token: string, transactionData: {
  amount_cents: number
  category_id?: number
  subcategory_id?: number
  operation_type: 'expense' | 'income'
  timestamp: string
  is_shared: boolean
  group_id?: number
}) => {
  const res = await axios.post(`${API_BASE}/transactions`, transactionData, {
    headers: { Authorization: `Bearer ${token}` }
  })
  // Clear relevant caches
  apiCache.clearPattern('/api/transactions')
  apiCache.clearPattern('/api/expenses')
  apiCache.clearPattern('/api/incomes')
  apiCache.clearPattern('/api/balance')
  return res.data
}

