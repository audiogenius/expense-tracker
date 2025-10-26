import * as React from 'react'
import type { Transaction } from '../../../types'

type DeletedTransaction = Transaction & {
  deleted_at: string
}

type TransactionListProps = {
  transactions: (Transaction | DeletedTransaction)[]
  selectedTransaction: Transaction | null
  onSelectTransaction: (transaction: Transaction) => void
  onDelete: (transaction: Transaction | DeletedTransaction) => void
  getTransactionColor: (operationType: string) => string
  getTransactionIcon: (operationType: string) => string
  isDeleted?: boolean
}

export const TransactionList = ({
  transactions,
  selectedTransaction,
  onSelectTransaction,
  onDelete,
  getTransactionColor,
  getTransactionIcon,
  isDeleted = false
}: TransactionListProps) => {
  if (transactions.length === 0) {
    return (
      <div className="empty-state">
        <p>{isDeleted ? 'Нет удаленных операций' : 'Операции не найдены'}</p>
      </div>
    )
  }

  return (
    <div className="transactions-grid">
      {transactions.map((transaction: Transaction) => (
        <div 
          key={transaction.id} 
          className={`transaction-card ${selectedTransaction?.id === transaction.id ? 'selected' : ''} ${isDeleted ? 'deleted' : ''}`}
          onClick={() => onSelectTransaction(transaction)}
        >
          <div className="transaction-icon" style={{ color: getTransactionColor(transaction.operation_type) }}>
            {getTransactionIcon(transaction.operation_type)}
          </div>
          
          <div className="transaction-details">
            <div className="transaction-amount" style={{ color: getTransactionColor(transaction.operation_type) }}>
              {transaction.operation_type === 'income' ? '+' : '-'}₽{Math.abs(transaction.amount_cents / 100).toLocaleString()}
            </div>
            <div className="transaction-category">
              {transaction.category_name || 'Без категории'}
            </div>
            <div className="transaction-date">
              {new Date(transaction.timestamp).toLocaleDateString('ru-RU')}
            </div>
            {isDeleted && 'deleted_at' in transaction && (
              <div className="deleted-date">
                Удалено: {new Date(transaction.deleted_at).toLocaleDateString('ru-RU')}
              </div>
            )}
          </div>
          
          <div className="transaction-actions">
            <button 
              className="delete-btn"
              onClick={(e) => {
                e.stopPropagation()
                onDelete(transaction)
              }}
            >
              {isDeleted ? 'Восстановить' : 'Удалить'}
            </button>
          </div>
        </div>
      ))}
    </div>
  )
}
