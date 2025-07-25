#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Shutting down servers...${NC}"
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null
    exit 0
}

# Set trap for cleanup
trap cleanup INT TERM

echo -e "${GREEN}Starting Pictoria Development Servers${NC}"
echo "======================================"

# Start backend
echo -e "${YELLOW}Starting backend on http://localhost:8080...${NC}"
go run . &
BACKEND_PID=$!

# Wait a bit for backend to start
sleep 2

# Start frontend
echo -e "${YELLOW}Starting frontend on http://localhost:5173...${NC}"
cd frontend && npm run dev &
FRONTEND_PID=$!

# Wait a bit for frontend to start
sleep 3

echo -e "\n${GREEN}✓ Both servers are running!${NC}"
echo -e "${GREEN}✓ Frontend: http://localhost:5173${NC}"
echo -e "${GREEN}✓ Backend API: http://localhost:8080${NC}"
echo -e "\n${YELLOW}Press Ctrl+C to stop both servers${NC}\n"

# Wait for both processes
wait $BACKEND_PID $FRONTEND_PID