export const getCategoryName = (categoryId: number | null | undefined, categories: any[]) => {
  if (!categoryId) return 'Без категории'
  const cat = categories.find((c: any) => c.id === categoryId)
  return cat ? cat.name : 'Неизвестно'
}

export const getIncomeTypeName = (type: string): string => {
  const types: Record<string, string> = {
    'salary': 'Зарплата',
    'debt_return': 'Возврат долга',
    'prize': 'Выигрыш',
    'gift': 'Подарок',
    'refund': 'Возврат средств',
    'other': 'Прочее'
  }
  return types[type] || 'Неизвестно'
}

export const formatDate = (timestamp: string): string => {
  return new Date(timestamp).toLocaleString('ru-RU')
}

export const formatCurrency = (cents: number): string => {
  return (cents / 100).toFixed(2) + ' ₽'
}

