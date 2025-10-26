import { useState, useEffect } from 'react'
import { 
  fetchCategories, 
  fetchSubcategories, 
  createCategory, 
  updateCategory, 
  deleteCategory,
  createSubcategory,
  updateSubcategory,
  deleteSubcategory
} from '../../../api'
import type { Category, Subcategory } from '../../../types'

export const useCategories = (token: string) => {
  const [categories, setCategories] = useState<Category[]>([])
  const [subcategories, setSubcategories] = useState<Subcategory[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

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

  const handleCreateCategory = async (name: string) => {
    try {
      setError(null)
      await createCategory(token, name.trim())
      await loadCategories()
    } catch (error: any) {
      setError(error.message || 'Ошибка создания категории')
    }
  }

  const handleUpdateCategory = async (category: Category, name: string) => {
    try {
      setError(null)
      await updateCategory(token, category.id, name.trim())
      await loadCategories()
    } catch (error: any) {
      setError(error.message || 'Ошибка обновления категории')
    }
  }

  const handleDeleteCategory = async (category: Category) => {
    try {
      setError(null)
      await deleteCategory(token, category.id)
      await loadCategories()
    } catch (error: any) {
      setError(error.message || 'Ошибка удаления категории')
    }
  }

  const handleCreateSubcategory = async (name: string, categoryId: number) => {
    try {
      setError(null)
      await createSubcategory(token, name.trim(), categoryId)
      await loadSubcategories(categoryId)
    } catch (error: any) {
      setError(error.message || 'Ошибка создания подкатегории')
    }
  }

  const handleUpdateSubcategory = async (subcategory: Subcategory, name: string) => {
    try {
      setError(null)
      await updateSubcategory(token, subcategory.id, name.trim(), subcategory.category_id)
      await loadSubcategories(subcategory.category_id)
    } catch (error: any) {
      setError(error.message || 'Ошибка обновления подкатегории')
    }
  }

  const handleDeleteSubcategory = async (subcategory: Subcategory) => {
    try {
      setError(null)
      await deleteSubcategory(token, subcategory.id)
      await loadSubcategories(subcategory.category_id)
    } catch (error: any) {
      setError(error.message || 'Ошибка удаления подкатегории')
    }
  }

  useEffect(() => {
    loadCategories()
  }, [])

  return {
    categories,
    subcategories,
    loading,
    error,
    setError,
    loadCategories,
    loadSubcategories,
    handleCreateCategory,
    handleUpdateCategory,
    handleDeleteCategory,
    handleCreateSubcategory,
    handleUpdateSubcategory,
    handleDeleteSubcategory
  }
}
