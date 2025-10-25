import { useState, useEffect } from 'react'
import { addExpense, fetchCategories, fetchSubcategories } from '../../api'
import { CategoryAutocomplete } from '../Suggestions/CategoryAutocomplete'
import type { Category, Subcategory, CategorySuggestion } from '../../types'

interface AddTransactionProps {
  token: string
  onTransactionAdded: () => void
}

export const AddTransaction = ({ token, onTransactionAdded }: AddTransactionProps) => {
  const [amount, setAmount] = useState('')
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(null)
  const [selectedSubcategory, setSelectedSubcategory] = useState<Subcategory | null>(null)
  const [categories, setCategories] = useState<Category[]>([])
  const [subcategories, setSubcategories] = useState<Subcategory[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadCategories()
  }, [])

  useEffect(() => {
    if (selectedCategory) {
      loadSubcategories(selectedCategory.id)
    } else {
      setSubcategories([])
      setSelectedSubcategory(null)
    }
  }, [selectedCategory])

  const loadCategories = async () => {
    try {
      const data = await fetchCategories()
      setCategories(data)
    } catch (error) {
      console.error('Failed to load categories:', error)
    }
  }

  const loadSubcategories = async (categoryId: number) => {
    try {
      const data = await fetchSubcategories(token, categoryId)
      setSubcategories(data)
    } catch (error) {
      console.error('Failed to load subcategories:', error)
    }
  }

  const handleSuggestionSelect = (suggestion: CategorySuggestion) => {
    if (suggestion.type === 'category') {
      const category = categories.find((c: Category) => c.id === suggestion.id)
      if (category) {
        setSelectedCategory(category)
        setSelectedSubcategory(null)
      }
    } else {
      const subcategory = subcategories.find((s: Subcategory) => s.id === suggestion.id)
      if (subcategory) {
        setSelectedSubcategory(subcategory)
        setSelectedCategory(categories.find((c: Category) => c.id === subcategory.category_id) || null)
      }
    }
    setSearchQuery('')
  }

  const handleSubmit = async (operationType: 'expense' | 'income') => {
    if (!amount || !selectedCategory) return

    try {
      setLoading(true)
      const amountCents = Math.round(parseFloat(amount) * 100)
      
      await addExpense(
        token,
        amountCents,
        selectedCategory.id,
        selectedSubcategory?.id || null,
        operationType
      )

      // Reset form
      setAmount('')
      setSelectedCategory(null)
      setSelectedSubcategory(null)
      setSearchQuery('')
      
      onTransactionAdded()
    } catch (error) {
      console.error('Failed to add transaction:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="glass-card add-transaction">
      <h3>Добавить операцию</h3>
      
      <div className="form-group">
        <label htmlFor="amount">Сумма (₽)</label>
        <input
          id="amount"
          type="number"
          step="0.01"
          min="0"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="0.00"
          className="form-input"
        />
      </div>

      <div className="form-group">
        <label htmlFor="category-search">Категория</label>
        <CategoryAutocomplete
          value={searchQuery}
          onChange={setSearchQuery}
          onSelect={handleSuggestionSelect}
          placeholder="Начните вводить название категории..."
          token={token}
        />
      </div>

      {selectedCategory && (
        <div className="form-group">
          <label>Выбранная категория</label>
          <div className="selected-category">
            <span className="category-name">{selectedCategory.name}</span>
            <button
              type="button"
              onClick={() => {
                setSelectedCategory(null)
                setSelectedSubcategory(null)
              }}
              className="clear-btn"
            >
              ✕
            </button>
          </div>
        </div>
      )}

      {selectedCategory && subcategories.length > 0 && (
        <div className="form-group">
          <label htmlFor="subcategory">Подкатегория (опционально)</label>
          <select
            id="subcategory"
            value={selectedSubcategory?.id || ''}
            onChange={(e) => {
              const subcategory = subcategories.find((s: Subcategory) => s.id === parseInt(e.target.value))
              setSelectedSubcategory(subcategory || null)
            }}
            className="form-select"
          >
            <option value="">Выберите подкатегорию</option>
            {subcategories.map((subcategory: Subcategory) => (
              <option key={subcategory.id} value={subcategory.id}>
                {subcategory.name}
              </option>
            ))}
          </select>
        </div>
      )}

      <div className="transaction-buttons">
        <button
          type="button"
          onClick={() => handleSubmit('expense')}
          disabled={!amount || !selectedCategory || loading}
          className="btn btn-expense"
        >
          {loading ? 'Добавление...' : 'Добавить расход'}
        </button>
        <button
          type="button"
          onClick={() => handleSubmit('income')}
          disabled={!amount || !selectedCategory || loading}
          className="btn btn-income"
        >
          {loading ? 'Добавление...' : 'Добавить приход'}
        </button>
      </div>
    </div>
  )
}