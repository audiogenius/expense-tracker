import * as React from 'react'
import type { Category, Subcategory } from '../../types'
import { useCategories } from './hooks/useCategories'
import { CategoryCard } from './components/CategoryCard'
import { SubcategoryCard } from './components/SubcategoryCard'
import { CategoryForm } from './components/CategoryForm'

type CategoriesPageProps = {
  token: string
  editable?: boolean
}

export const CategoriesPage = ({ token, editable = false }: CategoriesPageProps) => {
  const {
    categories,
    subcategories,
    loading,
    error,
    setError,
    loadSubcategories,
    handleCreateCategory,
    handleUpdateCategory,
    handleDeleteCategory,
    handleCreateSubcategory,
    handleUpdateSubcategory,
    handleDeleteSubcategory
  } = useCategories(token)

  const [selectedCategory, setSelectedCategory] = React.useState<Category | null>(null)
  const [showAddCategory, setShowAddCategory] = React.useState(false)
  const [showAddSubcategory, setShowAddSubcategory] = React.useState(false)
  const [editingCategory, setEditingCategory] = React.useState<Category | null>(null)
  const [editingSubcategory, setEditingSubcategory] = React.useState<Subcategory | null>(null)

  React.useEffect(() => {
    if (selectedCategory) {
      loadSubcategories(selectedCategory.id)
    }
  }, [selectedCategory, loadSubcategories])

  const handleCategorySelect = (category: Category) => {
    setSelectedCategory(category)
  }

  const handleCategoryEdit = (category: Category) => {
    setEditingCategory(category)
  }

  const handleCategoryUpdate = async (category: Category, name: string) => {
    await handleUpdateCategory(category, name)
    setEditingCategory(null)
  }

  const handleCategoryDelete = async (category: Category) => {
    if (!confirm(`Удалить категорию "${category.name}"?`)) return
    await handleDeleteCategory(category)
    if (selectedCategory?.id === category.id) {
      setSelectedCategory(null)
    }
  }

  const handleSubcategoryUpdate = async (subcategory: Subcategory, name: string) => {
    await handleUpdateSubcategory(subcategory, name)
    setEditingSubcategory(null)
  }

  const handleSubcategoryDelete = async (subcategory: Subcategory) => {
    if (!confirm(`Удалить подкатегорию "${subcategory.name}"?`)) return
    await handleDeleteSubcategory(subcategory)
  }

  const handleCreateCategorySubmit = async (name: string) => {
    await handleCreateCategory(name)
    setShowAddCategory(false)
  }

  const handleCreateSubcategorySubmit = async (name: string) => {
    if (!selectedCategory) return
    await handleCreateSubcategory(name, selectedCategory.id)
    setShowAddSubcategory(false)
  }

  if (loading) {
    return (
      <div className="categories-page">
        <div className="loading-indicator">
          <div className="spinner"></div>
          <span>Загрузка категорий...</span>
        </div>
      </div>
    )
  }

  return (
    <div className="categories-page">
      <div className="categories-header">
        <h1>Категории и подкатегории</h1>
        <p className="subtitle">
          {editable 
            ? 'Управление категориями и подкатегориями'
            : 'Справочник доступных категорий для расходов'
          }
        </p>
        {error && (
          <div className="error-message">
            ❌ {error}
            <button onClick={() => setError(null)}>✕</button>
          </div>
        )}
      </div>

      <div className="categories-grid">
        {/* Categories List */}
        <div className="categories-list-container">
          <div className="list-header">
            <h2>Категории ({categories.length})</h2>
            {editable && (
              <button className="add-btn" onClick={() => setShowAddCategory(true)}>
                ➕ Добавить
              </button>
            )}
          </div>
          <div className="categories-list">
            {showAddCategory && (
              <CategoryForm
                onSubmit={handleCreateCategorySubmit}
                onCancel={() => setShowAddCategory(false)}
                placeholder="Название категории"
              />
            )}

            {categories.map((category: Category) => (
              <CategoryCard
                key={category.id}
                category={category}
                isSelected={selectedCategory?.id === category.id}
                isEditing={editingCategory?.id === category.id}
                onSelect={handleCategorySelect}
                onEdit={handleCategoryEdit}
                onUpdate={handleCategoryUpdate}
                onDelete={handleCategoryDelete}
                onCancelEdit={() => setEditingCategory(null)}
                editable={editable}
              />
            ))}
          </div>
        </div>

        {/* Subcategories List */}
        <div className="subcategories-list-container">
          <div className="list-header">
            <h2>
              {selectedCategory 
                ? `Подкатегории для "${selectedCategory.name}" (${subcategories.length})`
                : 'Выберите категорию'}
            </h2>
            {editable && selectedCategory && (
              <button className="add-btn" onClick={() => setShowAddSubcategory(true)}>
                ➕ Добавить
              </button>
            )}
          </div>
          {selectedCategory ? (
            <div className="subcategories-list">
              {showAddSubcategory && (
                <CategoryForm
                  onSubmit={handleCreateSubcategorySubmit}
                  onCancel={() => setShowAddSubcategory(false)}
                  placeholder="Название подкатегории"
                />
              )}

              {subcategories.length === 0 ? (
                <div className="empty-state">
                  <p>Нет подкатегорий для этой категории</p>
                </div>
              ) : (
                subcategories.map((subcategory: Subcategory) => (
                  <SubcategoryCard
                    key={subcategory.id}
                    subcategory={subcategory}
                    isEditing={editingSubcategory?.id === subcategory.id}
                    onUpdate={handleSubcategoryUpdate}
                    onDelete={handleSubcategoryDelete}
                    onCancelEdit={() => setEditingSubcategory(null)}
                    editable={editable}
                  />
                ))
              )}
            </div>
          ) : (
            <div className="empty-state">
              <p>Выберите категорию слева, чтобы увидеть её подкатегории</p>
            </div>
          )}
        </div>
      </div>

      <div className="categories-info">
        <h3>ℹ️ Информация</h3>
        <ul>
          <li>Категории используются для классификации расходов</li>
          <li>Подкатегории позволяют детализировать расходы внутри категории</li>
          <li>При добавлении расхода можно выбрать категорию и подкатегорию</li>
          <li>Автокомплит подсказывает категории на основе вашей истории</li>
        </ul>
      </div>
    </div>
  )
}

