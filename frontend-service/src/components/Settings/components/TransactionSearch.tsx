import * as React from 'react'

type TransactionSearchProps = {
  searchQuery: string
  setSearchQuery: (query: string) => void
  onSearch: () => void
  loading: boolean
}

export const TransactionSearch = ({
  searchQuery,
  setSearchQuery,
  onSearch,
  loading
}: TransactionSearchProps) => {
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      onSearch()
    }
  }

  return (
    <div className="search-section">
      <div className="search-input-group">
        <input
          type="text"
          placeholder="Поиск операций..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onKeyPress={handleKeyPress}
          className="search-input"
        />
        <button 
          onClick={onSearch}
          disabled={loading || !searchQuery.trim()}
          className="search-btn"
        >
          {loading ? 'Поиск...' : 'Найти'}
        </button>
      </div>
    </div>
  )
}
