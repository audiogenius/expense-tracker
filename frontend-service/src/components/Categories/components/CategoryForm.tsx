import React, { useState } from 'react'
import type { Category } from '../../../types'

type CategoryFormProps = {
  category?: Category
  onSubmit: (name: string) => void
  onCancel: () => void
  placeholder: string
}

export const CategoryForm: React.FC<CategoryFormProps> = ({ 
  category, 
  onSubmit, 
  onCancel, 
  placeholder 
}) => {
  const [name, setName] = useState(category?.name || '')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (name.trim()) {
      onSubmit(name.trim())
      setName('')
    }
  }

  return (
    <div className="add-form">
      <input
        type="text"
        placeholder={placeholder}
        value={name}
        onChange={(e) => setName(e.target.value)}
        onKeyPress={(e) => e.key === 'Enter' && handleSubmit(e)}
        autoFocus
      />
      <div className="form-actions">
        <button onClick={handleSubmit} disabled={!name.trim()}>
          ✅ Сохранить
        </button>
        <button onClick={onCancel}>
          ❌ Отмена
        </button>
      </div>
    </div>
  )
}
