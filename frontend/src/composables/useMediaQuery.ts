import { ref, onMounted, onBeforeUnmount, type Ref } from 'vue'

// Reactive `matchMedia` — true while the query matches.
export function useMediaQuery(query: string): Ref<boolean> {
  const mql = window.matchMedia(query)
  const matches = ref(mql.matches)
  const update = (e: MediaQueryListEvent | MediaQueryList) => (matches.value = e.matches)

  onMounted(() => mql.addEventListener('change', update))
  onBeforeUnmount(() => mql.removeEventListener('change', update))

  return matches
}
