export interface Analytics {
  total_page_views: number
  total_add_to_carts: number
  total_purchases: number
  conversion_rate: number
  average_purchase_value: number
  max_purchase_value: number
  min_purchase_value: number
  top_product_id: string | null
}
