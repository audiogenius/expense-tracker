import * as React from 'react'
import { fetchTransactions, softDeleteTransaction, restoreTransaction, fetchDeletedTransactions } from '../../../api'
import type { Transaction, TransactionFilters } from '../../../types'

type DeletedTransaction = Transaction & {
  deleted_at: string
}

export const useTransactionManagement = (token: string) => {
  const [searchQuery, setSearchQuery] = React.useState('')
  const [transactions, setTransactions] = React.useState<Transaction[]>([])
  const [deletedTransactions, setDeletedTransactions] = React.useState<DeletedTransaction[]>([])
  const [loading, setLoading] = React.useState(false)
  const [selectedTransaction, setSelectedTransaction] = React.useState<Transaction | null>(null)
  const [activeTab, setActiveTab] = React.useState<'search' | 'deleted' | 'add'>('search')
  const [showAddForm, setShowAddForm] = React.useState(false)

  React.useEffect(() => {
    if (activeTab === 'deleted') {
      loadDeletedTransactions()
    }
  }, [activeTab])

  const handleSearch = async () => {
    if (!searchQuery.trim()) return

    try {
      setLoading(true)
      const filters: TransactionFilters = {
        period: 'all',
        startDate: '',
        endDate: '',
        scope: 'all'
      }
      
      const data = await fetchTransactions(token, filters)
      setTransactions(data)
    } catch (error) {
      console.error('Failed to search transactions:', error)
      alert('Не удалось найти операции')
    } finally {
      setLoading(false)
    }
  }

  const loadDeletedTransactions = async () => {
    try {
      setLoading(true)
      const data = await fetchDeletedTransactions(token)
      setDeletedTransactions(data)
    } catch (error) {
      console.error('Failed to load deleted transactions:', error)
      alert('Не удалось загрузить удаленные операции')
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (transaction: Transaction) => {
    if (!confirm('Вы уверены, что хотите удалить эту операцию?')) return

    try {
      await softDeleteTransaction(token, transaction.id)
      
      // Обновляем список
      setTransactions(transactions.filter((t: Transaction) => t.id !== transaction.id))
      setSelectedTransaction(null)
      
      alert('✅ Операция удалена')
    } catch (error) {
      console.error('Failed to delete transaction:', error)
      alert('Не удалось удалить операцию')
    }
  }

  const handleRestore = async (transaction: Transaction | DeletedTransaction) => {
    if (!confirm('Вы уверены, что хотите восстановить эту операцию?')) return

    try {
      await restoreTransaction(token, transaction.id)
      
      // Обновляем список
      setDeletedTransactions(deletedTransactions.filter((t: DeletedTransaction) => t.id !== transaction.id))
      
      alert('✅ Операция восстановлена')
    } catch (error) {
      console.error('Failed to restore transaction:', error)
      alert('Не удалось восстановить операцию')
    }
  }

  const handleTransactionAdded = () => {
    setShowAddForm(false)
    setActiveTab('search')
    // Можно добавить логику для обновления списка
  }

  const getTransactionColor = (operationType: string) => {
    return operationType === 'income' ? '#10b981' : '#ef4444'
  }

  const getTransactionIcon = (operationType: string) => {
    return operationType === 'income' ? '📈' : '📉'
  }

  return {
    // State
    searchQuery,
    setSearchQuery,
    transactions,
    deletedTransactions,
    loading,
    selectedTransaction,
    setSelectedTransaction,
    activeTab,
    setActiveTab,
    showAddForm,
    setShowAddForm,
    
    // Actions
    handleSearch,
    handleDelete,
    handleRestore,
    handleTransactionAdded,
    getTransactionColor,
    getTransactionIcon
  }
}
