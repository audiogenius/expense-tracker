import React, { useState, useEffect } from 'react'
import { 
  fetchCategories, 
  fetchSubcategories, 
  createCategory, 
  updateCategory, 
  deleteCategory,
  createSubcategory,
  updateSubcategory,
  deleteSubcategory
} from '../../api'
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
  const [showAddCategory, setShowAddCategory] = useState(false)
  const [showAddSubcategory, setShowAddSubcategory] = useState(false)
  const [editingCategory, setEditingCategory] = useState<Category | null>(null)
  const [editingSubcategory, setEditingSubcategory] = useState<Subcategory | null>(null)
  const [newCategoryName, setNewCategoryName] = useState('')
  const [newSubcategoryName, setNewSubcategoryName] = useState('')
  const [error, setError] = useState<string | null>(null)

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

  const handleCreateCategory = async () => {
    if (!newCategoryName.trim()) return
    
    try {
      setError(null)
      await createCategory(token, newCategoryName.trim())
      setNewCategoryName('')
      setShowAddCategory(false)
      await loadCategories()
    } catch (error: any) {
      setError(error.message || 'Ошибка создания категории')
    }
  }

  const handleUpdateCategory = async (category: Category) => {
    if (!newCategoryName.trim()) return
    
    try {
      setError(null)
      await updateCategory(token, category.id, newCategoryName.trim())
      setNewCategoryName('')
      setEditingCategory(null)
      await loadCategories()
    } catch (error: any) {
      setError(error.message || 'Ошибка обновления категории')
    }
  }

  const handleDeleteCategory = async (category: Category) => {
    if (!confirm(`Удалить категорию "${category.name}"?`)) return
    
    try {
      setError(null)
      await deleteCategory(token, category.id)
      if (selectedCategory?.id === category.id) {
        setSelectedCategory(null)
      }
      await loadCategories()
    } catch (error: any) {
      setError(error.message || 'Ошибка удаления категории')
    }
  }

  const handleCreateSubcategory = async () => {
    if (!newSubcategoryName.trim() || !selectedCategory) return
    
    try {
      setError(null)
      await createSubcategory(token, newSubcategoryName.trim(), selectedCategory.id)
      setNewSubcategoryName('')
      setShowAddSubcategory(false)
      await loadSubcategories(selectedCategory.id)
    } catch (error: any) {
      setError(error.message || 'Ошибка создания подкатегории')
    }
  }

  const handleUpdateSubcategory = async (subcategory: Subcategory) => {
    if (!newSubcategoryName.trim()) return
    
    try {
      setError(null)
      await updateSubcategory(token, subcategory.id, newSubcategoryName.trim(), subcategory.category_id)
      setNewSubcategoryName('')
      setEditingSubcategory(null)
      if (selectedCategory) {
        await loadSubcategories(selectedCategory.id)
      }
    } catch (error: any) {
      setError(error.message || 'Ошибка обновления подкатегории')
    }
  }

  const handleDeleteSubcategory = async (subcategory: Subcategory) => {
    if (!confirm(`Удалить подкатегорию "${subcategory.name}"?`)) return
    
    try {
      setError(null)
      await deleteSubcategory(token, subcategory.id)
      if (selectedCategory) {
        await loadSubcategories(selectedCategory.id)
      }
    } catch (error: any) {
      setError(error.message || 'Ошибка удаления подкатегории')
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
            {/* Add Category Form */}
            {showAddCategory && (
              <div className="add-form">
                <input
                  type="text"
                  placeholder="Название категории"
                  value={newCategoryName}
                  onChange={(e) => setNewCategoryName(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleCreateCategory()}
                  autoFocus
                />
                <div className="form-actions">
                  <button onClick={handleCreateCategory} disabled={!newCategoryName.trim()}>
                    ✅ Сохранить
                  </button>
                  <button onClick={() => {
                    setShowAddCategory(false)
                    setNewCategoryName('')
                  }}>
                    ❌ Отмена
                  </button>
                </div>
              </div>
            )}

            {categories.map((category) => (
              <div
                key={category.id}
                className={`category-card ${selectedCategory?.id === category.id ? 'selected' : ''}`}
                onClick={() => setSelectedCategory(category)}
              >
                {editingCategory?.id === category.id ? (
                  <div className="edit-form">
                    <input
                      type="text"
                      value={newCategoryName}
                      onChange={(e) => setNewCategoryName(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && handleUpdateCategory(category)}
                      autoFocus
                    />
                    <div className="form-actions">
                      <button onClick={() => handleUpdateCategory(category)} disabled={!newCategoryName.trim()}>
                        ✅ Сохранить
                      </button>
                      <button onClick={() => {
                        setEditingCategory(null)
                        setNewCategoryName('')
                      }}>
                        ❌ Отмена
                      </button>
                    </div>
                  </div>
                ) : (
                  <>
                    <div className="category-name">{category.name}</div>
                    <div className="category-meta">
                      ID: {category.id}
                      {editable && (
                        <div className="category-actions">
                          <button 
                            className="edit-btn-small" 
                            onClick={(e) => {
                              e.stopPropagation()
                              setEditingCategory(category)
                              setNewCategoryName(category.name)
                            }}
                            title="Редактировать категорию"
                          >
                            ✏️
                          </button>
                          <button 
                            className="delete-btn-small" 
                            onClick={(e) => {
                              e.stopPropagation()
                              handleDeleteCategory(category)
                            }}
                            title="Удалить категорию"
                          >
                            🗑️
                          </button>
                        </div>
                      )}
                    </div>
                  </>
                )}
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
              <button className="add-btn" onClick={() => setShowAddSubcategory(true)}>
                ➕ Добавить
              </button>
            )}
          </div>
          {selectedCategory ? (
            <div className="subcategories-list">
              {/* Add Subcategory Form */}
              {showAddSubcategory && (
                <div className="add-form">
                  <input
                    type="text"
                    placeholder="Название подкатегории"
                    value={newSubcategoryName}
                    onChange={(e) => setNewSubcategoryName(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleCreateSubcategory()}
                    autoFocus
                  />
                  <div className="form-actions">
                    <button onClick={handleCreateSubcategory} disabled={!newSubcategoryName.trim()}>
                      ✅ Сохранить
                    </button>
                    <button onClick={() => {
                      setShowAddSubcategory(false)
                      setNewSubcategoryName('')
                    }}>
                      ❌ Отмена
                    </button>
                  </div>
                </div>
              )}

              {subcategories.length === 0 ? (
                <div className="empty-state">
                  <p>Нет подкатегорий для этой категории</p>
                </div>
              ) : (
                subcategories.map((subcategory) => (
                  <div key={subcategory.id} className="subcategory-card">
                    {editingSubcategory?.id === subcategory.id ? (
                      <div className="edit-form">
                        <input
                          type="text"
                          value={newSubcategoryName}
                          onChange={(e) => setNewSubcategoryName(e.target.value)}
                          onKeyPress={(e) => e.key === 'Enter' && handleUpdateSubcategory(subcategory)}
                          autoFocus
                        />
                        <div className="form-actions">
                          <button onClick={() => handleUpdateSubcategory(subcategory)} disabled={!newSubcategoryName.trim()}>
                            ✅ Сохранить
                          </button>
                          <button onClick={() => {
                            setEditingSubcategory(null)
                            setNewSubcategoryName('')
                          }}>
                            ❌ Отмена
                          </button>
                        </div>
                      </div>
                    ) : (
                      <>
                        <div className="subcategory-name">{subcategory.name}</div>
                        <div className="subcategory-meta">
                          ID: {subcategory.id}
                          {editable && (
                            <div className="subcategory-actions">
                              <button 
                                className="edit-btn-small" 
                                onClick={(e) => {
                                  e.stopPropagation()
                                  setEditingSubcategory(subcategory)
                                  setNewSubcategoryName(subcategory.name)
                                }}
                                title="Редактировать подкатегорию"
                              >
                                ✏️
                              </button>
                              <button 
                                className="delete-btn-small" 
                                onClick={(e) => {
                                  e.stopPropagation()
                                  handleDeleteSubcategory(subcategory)
                                }}
                                title="Удалить подкатегорию"
                              >
                                🗑️
                              </button>
                            </div>
                          )}
                        </div>
                      </>
                    )}
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

