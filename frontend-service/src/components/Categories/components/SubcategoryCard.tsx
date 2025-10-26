import React from 'react'
import type { Subcategory } from '../../../types'
import { CategoryForm } from './CategoryForm'

type SubcategoryCardProps = {
  subcategory: Subcategory
  isEditing: boolean
  onUpdate: (subcategory: Subcategory, name: string) => void
  onDelete: (subcategory: Subcategory) => void
  onCancelEdit: () => void
  editable: boolean
}

export const SubcategoryCard: React.FC<SubcategoryCardProps> = ({
  subcategory,
  isEditing,
  onUpdate,
  onDelete,
  onCancelEdit,
  editable
}) => {
  if (isEditing) {
    return (
      <div className="subcategory-card">
        <CategoryForm
          category={subcategory}
          onSubmit={(name) => onUpdate(subcategory, name)}
          onCancel={onCancelEdit}
          placeholder="Название подкатегории"
        />
      </div>
    )
  }

  return (
    <div className="subcategory-card">
      <div className="subcategory-name">{subcategory.name}</div>
      <div className="subcategory-meta">
        ID: {subcategory.id}
        {editable && (
          <div className="subcategory-actions">
            <button 
              className="edit-btn-small" 
              onClick={(e) => {
                e.stopPropagation()
                onUpdate(subcategory, subcategory.name)
              }}
              title="Редактировать подкатегорию"
            >
              ✏️
            </button>
            <button 
              className="delete-btn-small" 
              onClick={(e) => {
                e.stopPropagation()
                onDelete(subcategory)
              }}
              title="Удалить подкатегорию"
            >
              🗑️
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
