import React, { useState, useEffect } from 'react'
import { fetchTransactions, softDeleteTransaction, restoreTransaction, fetchDeletedTransactions } from '../../api'
import type { Transaction, TransactionFilters } from '../../types'
import { formatCurrency, formatDate } from '../../utils/helpers'
import { AddTransactionForm } from './AddTransactionForm'

type TransactionManagementProps = {
  token: string
}

type DeletedTransaction = Transaction & {
  deleted_at: string
}

export const TransactionManagement: React.FC<TransactionManagementProps> = ({ token }) => {
  const [searchQuery, setSearchQuery] = useState('')
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [deletedTransactions, setDeletedTransactions] = useState<DeletedTransaction[]>([])
  const [loading, setLoading] = useState(false)
  const [selectedTransaction, setSelectedTransaction] = useState<Transaction | null>(null)
  const [activeTab, setActiveTab] = useState<'search' | 'deleted' | 'add'>('search')
  const [showAddForm, setShowAddForm] = useState(false)

  useEffect(() => {
    if (activeTab === 'deleted') {
      loadDeletedTransactions()
    }
  }, [activeTab])

  const handleSearch = async () => {
    if (!searchQuery.trim()) return

    try {
      setLoading(true)
      const filters: TransactionFilters = {
        operation_type: 'both',
        limit: 50
      }
      
      const response = await fetchTransactions(token, filters)
      
      // Фильтруем по поисковому запросу (категория, сумма, дата)
      const filtered = (response.transactions || []).filter((t) => {
        const searchLower = searchQuery.toLowerCase()
        const categoryMatch = t.category_name?.toLowerCase().includes(searchLower)
        const subcategoryMatch = t.subcategory_name?.toLowerCase().includes(searchLower)
        const amountMatch = (t.amount_cents / 100).toString().includes(searchQuery)
        const dateMatch = formatDate(t.timestamp).toLowerCase().includes(searchLower)
        
        return categoryMatch || subcategoryMatch || amountMatch || dateMatch
      })
      
      setTransactions(filtered)
    } catch (error) {
      console.error('Failed to search transactions:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadDeletedTransactions = async () => {
    try {
      setLoading(true)
      const response = await fetchDeletedTransactions(token)
      setDeletedTransactions(response.transactions || [])
    } catch (error) {
      console.error('Failed to load deleted transactions:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (transaction: Transaction) => {
    if (!confirm(`Удалить операцию на ${formatCurrency(transaction.amount_cents)}?`)) {
      return
    }

    try {
      await softDeleteTransaction(token, transaction.id)
      
      // Обновляем список
      setTransactions(transactions.filter(t => t.id !== transaction.id))
      setSelectedTransaction(null)
      
      alert('✅ Операция удалена')
    } catch (error) {
      console.error('Failed to delete transaction:', error)
      alert('Не удалось удалить операцию')
    }
  }

  const handleRestore = async (transaction: DeletedTransaction) => {
    if (!confirm(`Восстановить операцию на ${formatCurrency(transaction.amount_cents)}?`)) {
      return
    }

    try {
      await restoreTransaction(token, transaction.id)
      
      // Обновляем список
      setDeletedTransactions(deletedTransactions.filter(t => t.id !== transaction.id))
      
      alert('✅ Операция восстановлена')
    } catch (error) {
      console.error('Failed to restore transaction:', error)
      alert('Не удалось восстановить операцию')
    }
  }

  const handleTransactionAdded = () => {
    setShowAddForm(false)
    setActiveTab('search')
    // Optionally refresh the search results
    if (searchQuery.trim()) {
      handleSearch()
    }
  }

  const getTransactionColor = (operationType: 'expense' | 'income') => {
    return operationType === 'expense' ? 'var(--error)' : 'var(--success)'
  }

  const getTransactionIcon = (operationType: 'expense' | 'income') => {
    return operationType === 'expense' ? '↓' : '↑'
  }

  return (
    <div className="transaction-management">
      <div className="management-header">
        <h2>🗑️ Управление операциями</h2>
        <p className="subtitle">
          Найдите операцию по категории, сумме или дате и удалите её при необходимости
        </p>
      </div>

      <div className="management-tabs">
        <button 
          className={`tab ${activeTab === 'search' ? 'active' : ''}`}
          onClick={() => setActiveTab('search')}
        >
          🔍 Поиск операций
        </button>
        <button 
          className={`tab ${activeTab === 'add' ? 'active' : ''}`}
          onClick={() => setActiveTab('add')}
        >
          ➕ Добавить операцию
        </button>
        <button 
          className={`tab ${activeTab === 'deleted' ? 'active' : ''}`}
          onClick={() => setActiveTab('deleted')}
        >
          🗑️ Удаленные ({deletedTransactions.length})
        </button>
      </div>

      {activeTab === 'search' && (
        <div className="search-section">
          <div className="search-input-group">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
              placeholder="Поиск по категории, сумме или дате..."
              className="search-input"
            />
            <button 
              onClick={handleSearch} 
              disabled={loading || !searchQuery.trim()}
              className="search-btn"
            >
              {loading ? '⏳' : '🔍'} Найти
            </button>
          </div>
        </div>
      )}

      {activeTab === 'search' && transactions.length > 0 && (
        <div className="search-results">
          <div className="results-header">
            <h3>Найдено операций: {transactions.length}</h3>
          </div>
          
          <div className="transactions-grid">
            {transactions.map((transaction) => (
              <div 
                key={transaction.id} 
                className={`transaction-card ${selectedTransaction?.id === transaction.id ? 'selected' : ''}`}
                onClick={() => setSelectedTransaction(transaction)}
              >
                <div className="transaction-icon" style={{ color: getTransactionColor(transaction.operation_type) }}>
                  {getTransactionIcon(transaction.operation_type)}
                </div>
                
                <div className="transaction-details">
                  <div className="transaction-amount" style={{ color: getTransactionColor(transaction.operation_type) }}>
                    {transaction.operation_type === 'expense' ? '-' : '+'}{formatCurrency(transaction.amount_cents)}
                  </div>
                  <div className="transaction-info">
                    <div className="transaction-category">
                      {transaction.subcategory_name || transaction.category_name || 'Без категории'}
                    </div>
                    <div className="transaction-meta">
                      <span className="transaction-user">{transaction.username}</span>
                      <span className="transaction-date">{formatDate(transaction.timestamp)}</span>
                    </div>
                  </div>
                </div>

                <div className="transaction-actions">
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      handleDelete(transaction)
                    }}
                    className="delete-btn"
                    title="Удалить операцию"
                  >
                    🗑️
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {activeTab === 'add' && (
        <div className="add-section">
          <AddTransactionForm 
            token={token}
            onTransactionAdded={handleTransactionAdded}
            onCancel={() => setActiveTab('search')}
          />
        </div>
      )}

      {activeTab === 'deleted' && (
        <div className="deleted-transactions">
          <div className="results-header">
            <h3>Удаленные операции: {deletedTransactions.length}</h3>
            <button 
              onClick={loadDeletedTransactions}
              className="refresh-btn"
              disabled={loading}
            >
              {loading ? '⏳' : '🔄'} Обновить
            </button>
          </div>
          
          {deletedTransactions.length > 0 ? (
            <div className="transactions-grid">
              {deletedTransactions.map((transaction) => (
                <div 
                  key={transaction.id} 
                  className="transaction-card deleted"
                >
                  <div className="transaction-icon" style={{ color: getTransactionColor(transaction.operation_type) }}>
                    {getTransactionIcon(transaction.operation_type)}
                  </div>
                  
                  <div className="transaction-details">
                    <div className="transaction-amount" style={{ color: getTransactionColor(transaction.operation_type) }}>
                      {transaction.operation_type === 'expense' ? '-' : '+'}{formatCurrency(transaction.amount_cents)}
                    </div>
                    <div className="transaction-info">
                      <div className="transaction-category">
                        {transaction.subcategory_name || transaction.category_name || 'Без категории'}
                      </div>
                      <div className="transaction-meta">
                        <span className="transaction-user">{transaction.username}</span>
                        <span className="transaction-date">{formatDate(transaction.timestamp)}</span>
                        <span className="deleted-date">Удалено: {formatDate(transaction.deleted_at)}</span>
                      </div>
                    </div>
                  </div>

                  <div className="transaction-actions">
                    <button
                      onClick={() => handleRestore(transaction)}
                      className="restore-btn"
                      title="Восстановить операцию"
                    >
                      ♻️
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="empty-state">
              <div className="empty-state-icon">🗑️</div>
              <h3>Нет удаленных операций</h3>
              <p>Удаленные операции будут отображаться здесь</p>
            </div>
          )}
        </div>
      )}

      {activeTab === 'search' && transactions.length === 0 && searchQuery && !loading && (
        <div className="empty-state">
          <div className="empty-state-icon">🔍</div>
          <h3>Ничего не найдено</h3>
          <p>Попробуйте изменить поисковый запрос</p>
        </div>
      )}

      {activeTab === 'search' && !searchQuery && (
        <div className="info-box">
          <h4>💡 Как использовать:</h4>
          <ul>
            <li>Введите категорию (например, "Продукты")</li>
            <li>Или сумму (например, "100")</li>
            <li>Или дату (например, "26.10")</li>
            <li>Нажмите "Найти" или Enter</li>
            <li>Кликните на операцию и нажмите 🗑️ для удаления</li>
          </ul>
          
          <div className="info-box">
            <strong>✅ Мягкое удаление:</strong> Удаленные операции можно восстановить 
            во вкладке "Удаленные". Они не отображаются в основных списках.
          </div>
        </div>
      )}
    </div>
  )
}

