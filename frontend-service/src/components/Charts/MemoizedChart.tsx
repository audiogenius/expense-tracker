import React, { memo, useMemo, useRef, useEffect, useState } from 'react'
import { Line, Bar, Doughnut } from 'react-chartjs-2'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement,
} from 'chart.js'

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
)

interface ChartData {
  labels: string[]
  datasets: {
    label: string
    data: number[]
    backgroundColor?: string | string[]
    borderColor?: string | string[]
    borderWidth?: number
  }[]
}

interface MemoizedChartProps {
  type: 'line' | 'bar' | 'doughnut'
  data: ChartData
  title?: string
  height?: number
  options?: any
  className?: string
}

// Memoized chart component to prevent unnecessary re-renders
const MemoizedChart: React.FC<MemoizedChartProps> = memo(({
  type,
  data,
  title,
  height = 300,
  options,
  className = ''
}) => {
  const [isVisible, setIsVisible] = useState(false)
  const [observer, setObserver] = useState<IntersectionObserver | null>(null)

  // Memoized chart options
  const chartOptions = useMemo(() => ({
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'top' as const,
      },
      title: {
        display: !!title,
        text: title,
      },
    },
    scales: type !== 'doughnut' ? {
      y: {
        beginAtZero: true,
      },
    } : undefined,
    ...options,
  }), [title, options, type])

  // Memoized chart data
  const memoizedData = useMemo(() => data, [data])

  // Intersection Observer for lazy loading
  useEffect(() => {
    const chartElement = document.querySelector('.memoized-chart')
    if (!chartElement) return

    const obs = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setIsVisible(true)
            obs.unobserve(entry.target)
          }
        })
      },
      { threshold: 0.1 }
    )

    obs.observe(chartElement)
    setObserver(obs)

    return () => {
      if (obs) {
        obs.disconnect()
      }
    }
  }, [])

  // Cleanup observer on unmount
  useEffect(() => {
    return () => {
      if (observer) {
        observer.disconnect()
      }
    }
  }, [observer])

  const renderChart = () => {
    if (!isVisible) {
      return (
        <div 
          className="chart-placeholder"
          style={{ height: `${height}px` }}
        >
          <div className="loading-spinner"></div>
          <p>Загрузка графика...</p>
        </div>
      )
    }

    switch (type) {
      case 'line':
        return <Line data={memoizedData} options={chartOptions} />
      case 'bar':
        return <Bar data={memoizedData} options={chartOptions} />
      case 'doughnut':
        return <Doughnut data={memoizedData} options={chartOptions} />
      default:
        return null
    }
  }

  return (
    <div className={`memoized-chart ${className}`} style={{ height: `${height}px` }}>
      {renderChart()}
    </div>
  )
})

MemoizedChart.displayName = 'MemoizedChart'

export default MemoizedChart
