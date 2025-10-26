import * as React from 'react'
import { useTransactionManagement } from './hooks/useTransactionManagement'
import { TransactionSearch } from './components/TransactionSearch'
import { TransactionList } from './components/TransactionList'
import { AddTransactionForm } from './AddTransactionForm'

type TransactionManagementProps = {
  token: string
}

export const TransactionManagement = ({ token }: TransactionManagementProps) => {
  const {
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
    handleSearch,
    handleDelete,
    handleRestore,
    handleTransactionAdded,
    getTransactionColor,
    getTransactionIcon
  } = useTransactionManagement(token)

  return (
    <div className="transaction-management">
      <div className="management-tabs">
        <button 
          className={`tab ${activeTab === 'search' ? 'active' : ''}`}
          onClick={() => setActiveTab('search')}
        >
          🔍 Поиск операций
        </button>
        <button 
          className={`tab ${activeTab === 'deleted' ? 'active' : ''}`}
          onClick={() => setActiveTab('deleted')}
        >
          🗑️ Удаленные операции
        </button>
        <button 
          className={`tab ${activeTab === 'add' ? 'active' : ''}`}
          onClick={() => setActiveTab('add')}
        >
          ➕ Добавить операцию
        </button>
      </div>

      {activeTab === 'search' && (
        <div className="search-tab">
          <TransactionSearch
            searchQuery={searchQuery}
            setSearchQuery={setSearchQuery}
            onSearch={handleSearch}
            loading={loading}
          />
          
          {loading && (
            <div className="loading-indicator">
              <div className="spinner"></div>
              <p>Поиск операций...</p>
            </div>
          )}
          
          {!loading && transactions.length > 0 && (
            <div className="search-results">
              <div className="results-header">
                <h3>Найдено операций: {transactions.length}</h3>
              </div>
              
              <TransactionList
                transactions={transactions}
                selectedTransaction={selectedTransaction}
                onSelectTransaction={setSelectedTransaction}
                onDelete={handleDelete}
                getTransactionColor={getTransactionColor}
                getTransactionIcon={getTransactionIcon}
              />
            </div>
          )}
        </div>
      )}

      {activeTab === 'deleted' && (
        <div className="deleted-tab">
          <div className="deleted-header">
            <h3>Удаленные операции</h3>
            <button 
              className="refresh-btn"
              onClick={() => window.location.reload()}
            >
              🔄 Обновить
            </button>
          </div>
          
          {loading && (
            <div className="loading-indicator">
              <div className="spinner"></div>
              <p>Загрузка удаленных операций...</p>
            </div>
          )}
          
          <TransactionList
            transactions={deletedTransactions}
            selectedTransaction={selectedTransaction}
            onSelectTransaction={setSelectedTransaction}
            onDelete={(transaction) => handleRestore(transaction)}
            getTransactionColor={getTransactionColor}
            getTransactionIcon={getTransactionIcon}
            isDeleted={true}
          />
        </div>
      )}

      {activeTab === 'add' && (
        <div className="add-tab">
          <AddTransactionForm
            token={token}
            onTransactionAdded={handleTransactionAdded}
            onCancel={() => setActiveTab('search')}
          />
        </div>
      )}
    </div>
  )
}