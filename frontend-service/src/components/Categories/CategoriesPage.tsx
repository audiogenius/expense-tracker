import { useState, useEffect } from 'react'
import { fetchCategories, fetchSubcategories } from '../../api'
import type { Category, Subcategory } from '../../types'

type CategoriesPageProps = {
  token: string
  editable?: boolean
}

export const CategoriesPage: React.FC<CategoriesPageProps> = ({ token, editable = false }) => {
  const [categories, setCategories] = useState<Category[]>([])
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(null)
  const [subcategories, setSubcategories] = useState<Subcategory[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadCategories()
  }, [])

  useEffect(() => {
    if (selectedCategory) {
      loadSubcategories(selectedCategory.id)
    } else {
      setSubcategories([])
    }
  }, [selectedCategory])

  const loadCategories = async () => {
    try {
      setLoading(true)
      const data = await fetchCategories()
      setCategories(data)
    } catch (error) {
      console.error('Failed to load categories:', error)
    } finally {
      setLoading(false)
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
      </div>

      <div className="categories-grid">
        {/* Categories List */}
        <div className="categories-list-container">
          <div className="list-header">
            <h2>Категории ({categories.length})</h2>
            {editable && (
              <button className="add-btn" onClick={() => alert('⚠️ Функция добавления категорий будет доступна в следующей версии')}>
                ➕ Добавить
              </button>
            )}
          </div>
          <div className="categories-list">
            {categories.map((category) => (
              <div
                key={category.id}
                className={`category-card ${selectedCategory?.id === category.id ? 'selected' : ''}`}
                onClick={() => setSelectedCategory(category)}
              >
                <div className="category-name">{category.name}</div>
                <div className="category-meta">
                  ID: {category.id}
                  {editable && (
                    <button 
                      className="delete-btn-small" 
                      onClick={(e) => {
                        e.stopPropagation()
                        alert('⚠️ Функция удаления категорий будет доступна в следующей версии')
                      }}
                      title="Удалить категорию"
                    >
                      🗑️
                    </button>
                  )}
                </div>
              </div>
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
              <button className="add-btn" onClick={() => alert('⚠️ Функция добавления подкатегорий будет доступна в следующей версии')}>
                ➕ Добавить
              </button>
            )}
          </div>
          {selectedCategory ? (
            <div className="subcategories-list">
              {subcategories.length === 0 ? (
                <div className="empty-state">
                  <p>Нет подкатегорий для этой категории</p>
                </div>
              ) : (
                subcategories.map((subcategory) => (
                  <div key={subcategory.id} className="subcategory-card">
                    <div className="subcategory-name">{subcategory.name}</div>
                    <div className="subcategory-meta">
                      ID: {subcategory.id}
                      {editable && (
                        <button 
                          className="delete-btn-small" 
                          onClick={(e) => {
                            e.stopPropagation()
                            alert('⚠️ Функция удаления подкатегорий будет доступна в следующей версии')
                          }}
                          title="Удалить подкатегорию"
                        >
                          🗑️
                        </button>
                      )}
                    </div>
                  </div>
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

