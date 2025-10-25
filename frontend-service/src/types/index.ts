export type Expense = {
  id: number
  user_id: number
  amount_cents: number
  category_id?: number | null
  subcategory_id?: number | null
  operation_type: 'expense' | 'income'
  timestamp: string
  is_shared: boolean
  username?: string
  category_name?: string
  subcategory_name?: string
}

export type Income = {
  id: number
  user_id: number
  amount_cents: number
  income_type: string
  description?: string
  related_debt_id?: number | null
  timestamp: string
  username?: string
}

export type Category = {
  id: number
  name: string
  aliases: string[]
}

export type Subcategory = {
  id: number
  name: string
  category_id: number
  category_name: string
  aliases: string[]
  created_at: string
}

export type Transaction = {
  id: number
  user_id: number
  amount_cents: number
  category_id?: number | null
  subcategory_id?: number | null
  operation_type: 'expense' | 'income'
  timestamp: string
  is_shared: boolean
  username: string
  category_name?: string
  subcategory_name?: string
}

export type TransactionFilters = {
  operation_type?: 'expense' | 'income' | 'both'
  category_id?: number
  subcategory_id?: number
  start_date?: string
  end_date?: string
  page?: number
  limit?: number
  cursor?: string // For keyset pagination
}

export type TransactionResponse = {
  transactions: Transaction[]
  pagination: {
    limit: number
    has_more: boolean
    next_cursor?: string
  }
  filters: Record<string, string>
}

export type CategorySuggestion = {
  id: number
  name: string
  type: 'category' | 'subcategory'
  score: number
  usage: number
}

export type Balance = {
  balance_cents: number
  balance_rubles: number
  total_incomes_cents: number
  total_incomes_rubles: number
  total_expenses_cents: number
  total_expenses_rubles: number
  period: string
}

export type Profile = {
  username?: string
  id?: string
  photo_url?: string
}

export type Period = 'all' | 'day' | 'week' | 'month' | 'custom'

export type ChartPeriod = 'days' | 'months' | 'years'

export type ChartType = 'expenses' | 'incomes' | 'both'

export type CustomPeriod = {
  start_date: string
  end_date: string
}

