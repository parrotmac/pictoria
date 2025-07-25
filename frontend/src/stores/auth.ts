import { ref, watch } from 'vue'

const key = `name`
const defaultValue = ``

const stored = localStorage.getItem(key)
export const nameStore = ref(stored ? stored : defaultValue)

watch(nameStore, (newVal) => {
  console.log(`name`, newVal)
  localStorage.setItem(key, newVal)
})
