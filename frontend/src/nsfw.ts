import { ref, watch } from 'vue'

const NSFW_KEY = 'nsfwMode'
const saved = localStorage.getItem(NSFW_KEY)
const nsfw = ref<'hide'|'show'|'only'>(saved === 'show' || saved === 'only' ? saved : 'hide')

watch(nsfw, v => {
  localStorage.setItem(NSFW_KEY, v)
})

export { nsfw }
