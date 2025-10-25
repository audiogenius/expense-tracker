import React, { useState, useEffect, useCallback, useRef } from 'react'
import type { CategorySuggestion } from '../../types'

interface CategoryAutocompleteProps {
  value: string
  onChange: (value: string) => void
  onSelect: (suggestion: CategorySuggestion) => void
  placeholder?: string
  className?: string
  token: string
}

export const CategoryAutocomplete: React.FC<CategoryAutocompleteProps> = ({
  value,
  onChange,
  onSelect,
  placeholder = 'Введите категорию...',
  className = '',
  token
}) => {
  const [suggestions, setSuggestions] = useState<CategorySuggestion[]>([])
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [highlightedIndex, setHighlightedIndex] = useState(-1)
  
  const inputRef = useRef<HTMLInputElement>(null)
  const suggestionsRef = useRef<HTMLDivElement>(null)
  const debounceRef = useRef<number | undefined>(undefined)

  // Debounced search function
  const searchSuggestions = useCallback(async (query: string) => {
    if (query.length < 2) {
      setSuggestions([])
      setIsOpen(false)
      return
    }

    setIsLoading(true)
    
    try {
      const response = await fetch(
        `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/suggestions/categories?query=${encodeURIComponent(query)}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        }
      )

      if (response.ok) {
        const data = await response.json()
        setSuggestions(data)
        setIsOpen(data.length > 0)
        setHighlightedIndex(-1)
      } else {
        console.error('Failed to fetch suggestions')
        setSuggestions([])
        setIsOpen(false)
      }
    } catch (error) {
      console.error('Error fetching suggestions:', error)
      setSuggestions([])
      setIsOpen(false)
    } finally {
      setIsLoading(false)
    }
  }, [token])

  // Handle input change with debounce
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value
    onChange(newValue)

    // Clear previous timeout
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }

    // Set new timeout
    debounceRef.current = window.setTimeout(() => {
      searchSuggestions(newValue)
    }, 300)
  }

  // Handle suggestion selection
  const handleSuggestionSelect = (suggestion: CategorySuggestion) => {
    onChange(suggestion.name)
    onSelect(suggestion)
    setIsOpen(false)
    setSuggestions([])
    setHighlightedIndex(-1)
  }

  // Handle keyboard navigation
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (!isOpen || suggestions.length === 0) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setHighlightedIndex(prev => 
          prev < suggestions.length - 1 ? prev + 1 : prev
        )
        break
      case 'ArrowUp':
        e.preventDefault()
        setHighlightedIndex(prev => prev > 0 ? prev - 1 : -1)
        break
      case 'Enter':
        e.preventDefault()
        if (highlightedIndex >= 0 && highlightedIndex < suggestions.length) {
          handleSuggestionSelect(suggestions[highlightedIndex])
        }
        break
      case 'Escape':
        setIsOpen(false)
        setHighlightedIndex(-1)
        break
    }
  }

  // Handle click outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        suggestionsRef.current && 
        !suggestionsRef.current.contains(event.target as Node) &&
        inputRef.current &&
        !inputRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false)
        setHighlightedIndex(-1)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // Highlight matching text
  const highlightText = (text: string, query: string) => {
    if (!query) return text
    
    const regex = new RegExp(`(${query})`, 'gi')
    const parts = text.split(regex)
    
    return parts.map((part, index) => 
      regex.test(part) ? (
        <mark key={index} className="highlight">{part}</mark>
      ) : part
    )
  }

  // Group suggestions by type
  const groupedSuggestions = suggestions.reduce((acc, suggestion) => {
    if (!acc[suggestion.type]) {
      acc[suggestion.type] = []
    }
    acc[suggestion.type].push(suggestion)
    return acc
  }, {} as Record<string, CategorySuggestion[]>)

  return (
    <div className={`autocomplete-container ${className}`}>
      <div className="autocomplete-input-wrapper">
        <input
          ref={inputRef}
          type="text"
          value={value}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          onFocus={() => {
            if (suggestions.length > 0) {
              setIsOpen(true)
            }
          }}
          placeholder={placeholder}
          className="autocomplete-input"
          autoComplete="off"
        />
        {isLoading && (
          <div className="autocomplete-loading">
            <div className="spinner"></div>
          </div>
        )}
      </div>

      {isOpen && suggestions.length > 0 && (
        <div ref={suggestionsRef} className="autocomplete-suggestions">
          {Object.entries(groupedSuggestions).map(([type, typeSuggestions]) => (
            <div key={type} className="suggestion-group">
              <div className="suggestion-group-header">
                {type === 'category' ? 'Категории' : 'Подкатегории'}
              </div>
              {typeSuggestions.map((suggestion, index) => {
                const globalIndex = suggestions.findIndex(s => s.id === suggestion.id)
                const isHighlighted = globalIndex === highlightedIndex
                
                return (
                  <div
                    key={suggestion.id}
                    className={`suggestion-item ${isHighlighted ? 'highlighted' : ''}`}
                    onClick={() => handleSuggestionSelect(suggestion)}
                    onMouseEnter={() => setHighlightedIndex(globalIndex)}
                  >
                    <div className="suggestion-name">
                      {highlightText(suggestion.name, value)}
                    </div>
                    <div className="suggestion-meta">
                      <span className="suggestion-type">
                        {suggestion.type === 'category' ? '📁' : '📂'}
                      </span>
                      {suggestion.usage && suggestion.usage > 0 && (
                        <span className="suggestion-usage">
                          {suggestion.usage} раз
                        </span>
                      )}
                      <span className="suggestion-score">
                        {Math.round(suggestion.score * 100)}%
                      </span>
                    </div>
                  </div>
                )
              })}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
