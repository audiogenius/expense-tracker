export type Expense = {
  id: number
  user_id: number
  amount_cents: number
  category_id?: number | null
  timestamp: string
  is_shared: boolean
  username?: string
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

export type Period = 'all' | 'week' | 'month'

