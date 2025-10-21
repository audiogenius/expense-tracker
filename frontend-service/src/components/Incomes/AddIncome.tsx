import React, { useState } from 'react'

type AddIncomeProps = {
  onAdd: (amount: string, type: string, description: string) => Promise<void>
}

export const AddIncome: React.FC<AddIncomeProps> = ({ onAdd }) => {
  const [incomeAmount, setIncomeAmount] = useState<string>('')
  const [incomeType, setIncomeType] = useState<string>('salary')
  const [incomeDescription, setIncomeDescription] = useState<string>('')

  const handleAdd = async () => {
    await onAdd(incomeAmount, incomeType, incomeDescription)
    setIncomeAmount('')
    setIncomeType('salary')
    setIncomeDescription('')
  }

  return (
    <div className="glass-card">
      <h3>Добавить приход</h3>
      <div className="controls">
        <input
          type="number"
          placeholder="Сумма (руб.)"
          value={incomeAmount}
          onChange={(e) => setIncomeAmount(e.target.value)}
        />
        <select value={incomeType} onChange={(e) => setIncomeType(e.target.value)}>
          <option value="salary">Зарплата</option>
          <option value="debt_return">Возврат долга</option>
          <option value="prize">Выигрыш</option>
          <option value="gift">Подарок</option>
          <option value="refund">Возврат средств</option>
          <option value="other">Прочее</option>
        </select>
        <input
          type="text"
          placeholder="Описание (необязательно)"
          value={incomeDescription}
          onChange={(e) => setIncomeDescription(e.target.value)}
        />
        <button onClick={handleAdd} style={{ background: '#10b981' }}>
          Добавить приход
        </button>
      </div>
    </div>
  )
}

