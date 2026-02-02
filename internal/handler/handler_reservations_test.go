package handler

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"sync"
// 	"testing"
// )

// func TestCreateReservation_RaceCondition(t *testing.T) {
//     // Setup handler with real dependencies or mocks
//     handler := setupTestHandler()
    
//     // Run multiple concurrent requests
//     var wg sync.WaitGroup
//     numRequests := 100
    
//     for i := 0; i < numRequests; i++ {
//         wg.Add(1)
//         go func(id int) {
//             defer wg.Done()
            
//             req := createTestReservationRequest(id)
//             w := httptest.NewRecorder()
            
//             handler.CreateReservation(w, req)
            
//             if w.Code != http.StatusCreated {
//                 t.Errorf("Request %d failed with status %d", id, w.Code)
//             }
//         }(i)
//     }
    
//     wg.Wait()
// }