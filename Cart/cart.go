package Cart

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

var client *redis.Client

type Product struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type Cart struct {
	Items map[string]int `json:"items"`
}

func AddToCart(w http.ResponseWriter, r *http.Request) {
	var product Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId := r.Header.Get("id")
	if userId == "" {
		http.Error(w, "用户id无效", http.StatusBadRequest)
		return
	}
	cart := Cart{
		Items: map[string]int{},
	}
	val, err := client.Get(context.Background(), "cart:"+userId).Result()
	if err == nil {
		err = json.Unmarshal([]byte(val), &cart)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != redis.Nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cart.Items[strconv.Itoa(product.Id)] += 1
	jsonData, err := json.Marshal(cart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = client.Set(context.Background(), "cart:"+userId, jsonData, 0).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "成功添加商品到购物车中")

}

func GetCartList(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["userId"]
	if userId == "" {
		http.Error(w, "用户id无效", http.StatusBadRequest)
		return
	}
	val, err := client.Get(context.Background(), "cart:"+userId).Result()
	if err == redis.Nil {
		http.Error(w, "购物车为空", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var cart Cart
	err = json.Unmarshal([]byte(val), &cart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var products []Product
	for productId, quantiy := range cart.Items {
		productVal, err := client.Get(context.Background(), "product:"+productId).Result()
		if err == redis.Nil {
			http.Error(w, "商品列表为空", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var product Product
		err = json.Unmarshal([]byte(productVal), &product)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		product.Price *= quantiy
		products = append(products, product)
	}
	jsonData, err := json.Marshal(products)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
