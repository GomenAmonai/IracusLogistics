import { useEffect, useRef, useState } from 'react'

// Reveal по входу во вьюпорт через IntersectionObserver (не scroll-listener).
// Возвращает ref на элемент и флаг видимости — секция набегает один раз.
export function useReveal<T extends HTMLElement>() {
  const ref = useRef<T>(null)
  const [isVisible, setVisible] = useState(false)

  useEffect(() => {
    const node = ref.current
    if (!node) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setVisible(true)
          observer.disconnect()
        }
      },
      { threshold: 0.15 },
    )

    observer.observe(node)
    return () => observer.disconnect()
  }, [])

  return { ref, isVisible }
}
