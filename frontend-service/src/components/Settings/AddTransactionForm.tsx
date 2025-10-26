import React, { useState, useEffect } from 'react'
import { createTransaction, fetchCategories, fetchSubcategories } from '../../api'
import type { Category, Subcategory } from '../../types'

type AddTransactionFormProps = {
  token: string
  onTransactionAdded: () => void
  onCancel: () => void
}

export const AddTransactionForm: React.FC<AddTransactionFormProps> = ({ 
  token, 
  onTransactionAdded, 
  onCancel 
}) => {
  const [categories, setCategories] = useState<Category[]>([])
  const [subcategories, setSubcategories] = useState<Subcategory[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  
  // Form state
  const [amount, setAmount] = useState('')
  const [operationType, setOperationType] = useState<'expense' | 'income'>('expense')
  const [selectedCategoryId, setSelectedCategoryId] = useState<number | null>(null)
  const [selectedSubcategoryId, setSelectedSubcategoryId] = useState<number | null>(null)
  const [timestamp, setTimestamp] = useState(new Date().toISOString().slice(0, 16))
  const [isShared, setIsShared] = useState(false)

  useEffect(() => {
    loadCategories()
  }, [])

  useEffect(() => {
    if (selectedCategoryId) {
      loadSubcategories(selectedCategoryId)
    } else {
      setSubcategories([])
      setSelectedSubcategoryId(null)
    }
  }, [selectedCategoryId])

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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!amount || parseFloat(amount) <= 0) {
      setError('Введите корректную сумму')
      return
    }

    if (operationType === 'expense' && !selectedCategoryId) {
      setError('Выберите категорию для расхода')
      return
    }

    try {
      setLoading(true)
      setError(null)

      const amountCents = Math.round(parseFloat(amount) * 100)
      const transactionData = {
        amount_cents: amountCents,
        category_id: selectedCategoryId || undefined,
        subcategory_id: selectedSubcategoryId || undefined,
        operation_type: operationType,
        timestamp: new Date(timestamp).toISOString(),
        is_shared: isShared,
        group_id: undefined // TODO: Add group selection
      }

      await createTransaction(token, transactionData)
      onTransactionAdded()
    } catch (error: any) {
      setError(error.message || 'Ошибка создания операции')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="add-transaction-form">
      <div className="form-header">
        <h3>➕ Добавить операцию</h3>
        <button onClick={onCancel} className="close-btn">✕</button>
      </div>

      {error && (
        <div className="error-message">
          ❌ {error}
          <button onClick={() => setError(null)}>✕</button>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Тип операции</label>
          <div className="radio-group">
            <label className="radio-label">
              <input
                type="radio"
                value="expense"
                checked={operationType === 'expense'}
                onChange={(e) => setOperationType(e.target.value as 'expense' | 'income')}
              />
              <span>💸 Расход</span>
            </label>
            <label className="radio-label">
              <input
                type="radio"
                value="income"
                checked={operationType === 'income'}
                onChange={(e) => setOperationType(e.target.value as 'expense' | 'income')}
              />
              <span>💰 Приход</span>
            </label>
          </div>
        </div>

        <div className="form-group">
          <label htmlFor="amount">Сумма (руб.)</label>
          <input
            id="amount"
            type="number"
            step="0.01"
            min="0"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder="0.00"
            required
          />
        </div>

        {operationType === 'expense' && (
          <>
            <div className="form-group">
              <label htmlFor="category">Категория</label>
              <select
                id="category"
                value={selectedCategoryId || ''}
                onChange={(e) => setSelectedCategoryId(e.target.value ? parseInt(e.target.value) : null)}
                required
              >
                <option value="">Выберите категорию</option>
                {categories.map((category) => (
                  <option key={category.id} value={category.id}>
                    {category.name}
                  </option>
                ))}
              </select>
            </div>

            {selectedCategoryId && subcategories.length > 0 && (
              <div className="form-group">
                <label htmlFor="subcategory">Подкатегория</label>
                <select
                  id="subcategory"
                  value={selectedSubcategoryId || ''}
                  onChange={(e) => setSelectedSubcategoryId(e.target.value ? parseInt(e.target.value) : null)}
                >
                  <option value="">Выберите подкатегорию (необязательно)</option>
                  {subcategories.map((subcategory) => (
                    <option key={subcategory.id} value={subcategory.id}>
                      {subcategory.name}
                    </option>
                  ))}
                </select>
              </div>
            )}
          </>
        )}

        <div className="form-group">
          <label htmlFor="timestamp">Дата и время</label>
          <input
            id="timestamp"
            type="datetime-local"
            value={timestamp}
            onChange={(e) => setTimestamp(e.target.value)}
            required
          />
        </div>

        <div className="form-group">
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={isShared}
              onChange={(e) => setIsShared(e.target.checked)}
            />
            <span>Семейная операция (видна всем участникам группы)</span>
          </label>
        </div>

        <div className="form-actions">
          <button type="button" onClick={onCancel} disabled={loading}>
            ❌ Отмена
          </button>
          <button type="submit" disabled={loading}>
            {loading ? '⏳ Создание...' : '✅ Создать'}
          </button>
        </div>
      </form>
    </div>
  )
}
