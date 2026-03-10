package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	cities := []string{}
	for city := range cafeList {
		cities = append(cities, city)
	}

	for _, city := range cities {
		requests := []struct {
			count int
			want  int
		}{
			{0, 0},
			{1, 1},
			{2, 2},
			{100, min(100, len(cafeList[city]))},
		}

		for _, v := range requests {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/cafe?city=%s&count=%d", city, v.count), nil)
			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			resp := strings.TrimSpace(response.Body.String())

			if resp == "" {
				require.Equal(t, v.want, 0,
					"Найдено неверное количество кафе с параметром: %d", v.count)
				continue
			}
			assert.Equal(t, v.want, len(strings.Split(resp, ",")),
				"Найдено неверное количество кафе с параметром: %d", v.count)
		}
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	city := "moscow"

	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}

	for _, v := range requests {
		searchToLower := strings.ToLower(v.search)

		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/cafe?city=%s&search=%s", city, v.search), nil)
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		resp := strings.TrimSpace(response.Body.String())

		if resp == "" {
			require.Equal(t, v.wantCount, 0,
				"Найдено неверное количество кафе с параметром %s", v.search)
			continue
		}
		assert.Equal(t, v.wantCount, len(strings.Split(resp, ",")),
			"Найдено неверное количество кафе с параметром %s", v.search)

		cafes := strings.Split(resp, ",")

		for _, cafe := range cafes {
			assert.True(t, strings.Contains(strings.ToLower(cafe), searchToLower),
				"Кафе с названием: %s не содержит: %s", cafe, searchToLower)
		}
	}
}
