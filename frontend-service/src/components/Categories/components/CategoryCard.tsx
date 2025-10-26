import React from 'react'
import type { Category } from '../../../types'
import { CategoryForm } from './CategoryForm'

type CategoryCardProps = {
  category: Category
  isSelected: boolean
  isEditing: boolean
  onSelect: (category: Category) => void
  onEdit: (category: Category) => void
  onUpdate: (category: Category, name: string) => void
  onDelete: (category: Category) => void
  onCancelEdit: () => void
  editable: boolean
}

export const CategoryCard: React.FC<CategoryCardProps> = ({
  category,
  isSelected,
  isEditing,
  onSelect,
  onEdit,
  onUpdate,
  onDelete,
  onCancelEdit,
  editable
}) => {
  if (isEditing) {
    return (
      <div className={`category-card ${isSelected ? 'selected' : ''}`}>
        <CategoryForm
          category={category}
          onSubmit={(name) => onUpdate(category, name)}
          onCancel={onCancelEdit}
          placeholder="Название категории"
        />
      </div>
    )
  }

  return (
    <div
      className={`category-card ${isSelected ? 'selected' : ''}`}
      onClick={() => onSelect(category)}
    >
      <div className="category-name">{category.name}</div>
      <div className="category-meta">
        ID: {category.id}
        {editable && (
          <div className="category-actions">
            <button 
              className="edit-btn-small" 
              onClick={(e) => {
                e.stopPropagation()
                onEdit(category)
              }}
              title="Редактировать категорию"
            >
              ✏️
            </button>
            <button 
              className="delete-btn-small" 
              onClick={(e) => {
                e.stopPropagation()
                onDelete(category)
              }}
              title="Удалить категорию"
            >
              🗑️
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
