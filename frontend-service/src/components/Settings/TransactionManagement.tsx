import { useState } from 'react'
import { fetchTransactions } from '../../api'
import type { Transaction, TransactionFilters } from '../../types'
import { formatCurrency, formatDate } from '../../utils/helpers'

type TransactionManagementProps = {
  token: string
}

export const TransactionManagement: React.FC<TransactionManagementProps> = ({ token }) => {
  const [searchQuery, setSearchQuery] = useState('')
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(false)
  const [selectedTransaction, setSelectedTransaction] = useState<Transaction | null>(null)

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

  const handleDelete = async (transaction: Transaction) => {
    if (!confirm(`Удалить операцию на ${formatCurrency(transaction.amount_cents)}?`)) {
      return
    }

    try {
      // TODO: Implement soft delete API endpoint
      // await fetch(`/api/transactions/${transaction.id}`, {
      //   method: 'DELETE',
      //   headers: { 'Authorization': `Bearer ${token}` }
      // })
      
      alert('⚠️ Функция удаления будет доступна в следующей версии')
      
      // Обновляем список
      setTransactions(transactions.filter(t => t.id !== transaction.id))
      setSelectedTransaction(null)
    } catch (error) {
      console.error('Failed to delete transaction:', error)
      alert('Не удалось удалить операцию')
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

      {transactions.length > 0 && (
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

      {transactions.length === 0 && searchQuery && !loading && (
        <div className="empty-state">
          <div className="empty-state-icon">🔍</div>
          <h3>Ничего не найдено</h3>
          <p>Попробуйте изменить поисковый запрос</p>
        </div>
      )}

      {!searchQuery && (
        <div className="info-box">
          <h4>💡 Как использовать:</h4>
          <ul>
            <li>Введите категорию (например, "Продукты")</li>
            <li>Или сумму (например, "100")</li>
            <li>Или дату (например, "26.10")</li>
            <li>Нажмите "Найти" или Enter</li>
            <li>Кликните на операцию и нажмите 🗑️ для удаления</li>
          </ul>
          
          <div className="warning-box">
            <strong>⚠️ Внимание:</strong> Удаление операций необратимо. 
            В будущих версиях будет добавлена возможность "мягкого удаления" 
            с возможностью восстановления.
          </div>
        </div>
      )}
    </div>
  )
}

