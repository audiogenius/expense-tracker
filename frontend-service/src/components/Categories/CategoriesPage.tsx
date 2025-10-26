import { useState, useEffect } from 'react'
import { fetchCategories, fetchSubcategories } from '../../api'
import type { Category, Subcategory } from '../../types'

type CategoriesPageProps = {
  token: string
}

export const CategoriesPage: React.FC<CategoriesPageProps> = ({ token }) => {
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
        <p className="subtitle">Справочник доступных категорий для расходов</p>
      </div>

      <div className="categories-grid">
        {/* Categories List */}
        <div className="categories-list-container">
          <h2>Категории ({categories.length})</h2>
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
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Subcategories List */}
        <div className="subcategories-list-container">
          <h2>
            {selectedCategory 
              ? `Подкатегории для "${selectedCategory.name}" (${subcategories.length})`
              : 'Выберите категорию'}
          </h2>
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

