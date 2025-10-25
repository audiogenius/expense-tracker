import React, { useState } from 'react'
import type { Category } from '../../types'

type AddExpenseProps = {
  categories: Category[]
  onAdd: (amount: string, categoryId: number | null) => Promise<void>
}

export const AddExpense: React.FC<AddExpenseProps> = ({ categories, onAdd }) => {
  const [amount, setAmount] = useState<string>('')
  const [selectedCategory, setSelectedCategory] = useState<number | null>(null)

  const handleAdd = async () => {
    await onAdd(amount, selectedCategory)
    setAmount('')
    setSelectedCategory(null)
  }

  return (
    <div className="glass-card">
      <h3>Добавить расход</h3>
      <div className="controls">
        <input
          type="number"
          placeholder="Сумма (руб.)"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
        />
        <select
          value={selectedCategory || ''}
          onChange={(e) => setSelectedCategory(e.target.value ? parseInt(e.target.value) : null)}
        >
          <option value="">Без категории</option>
          {categories.map((cat) => (
            <option key={cat.id} value={cat.id}>
              {cat.name}
            </option>
          ))}
        </select>
        <button onClick={handleAdd}>Добавить расход</button>
      </div>
    </div>
  )
}

