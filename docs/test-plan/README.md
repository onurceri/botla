# Botla Platform - Comprehensive Test Plan

## Overview
This directory contains detailed test plans for every component of the Botla chatbot platform.

---

## Test Plan Index

### 1. Authentication & User Management
- [1_1_User_Registration.md](./1.%20Authentication%20%26%20User%20Management/1_1_User_Registration.md)
- [1_2_User_Login.md](./1.%20Authentication%20%26%20User%20Management/1_2_User_Login.md)
- [1_3_Token_Management.md](./1.%20Authentication%20%26%20User%20Management/1_3_Token_Management.md)
- [1_4_Profile_Management.md](./1.%20Authentication%20%26%20User%20Management/1_4_Profile_Management.md)

### 2. Plan & Subscription Management
- [2_1_Free_Plan_Limits.md](./2.%20Plan%20%26%20Subscription%20Management/2_1_Free_Plan_Limits.md)
- [2_2_Pro_Plan_Features.md](./2.%20Plan%20%26%20Subscription%20Management/2_2_Pro_Plan_Features.md)
- [2_3_Ultra_Plan_Features.md](./2.%20Plan%20%26%20Subscription%20Management/2_3_Ultra_Plan_Features.md)
- [2_4_Plan_Enforcement_Security.md](./2.%20Plan%20%26%20Subscription%20Management/2_4_Plan_Enforcement_Security.md)
- [2_5_Usage_Tracking.md](./2.%20Plan%20%26%20Subscription%20Management/2_5_Usage_Tracking.md)

### 3. Chatbot Management
- [3_1_Chatbot_Creation.md](./3.%20Chatbot%20Management/3_1_Chatbot_Creation.md)
- [3_2_Chatbot_Retrieval.md](./3.%20Chatbot%20Management/3_2_Chatbot_Retrieval.md)
- [3_3_Chatbot_Update.md](./3.%20Chatbot%20Management/3_3_Chatbot_Update.md)
- [3_4_Chatbot_Deletion.md](./3.%20Chatbot%20Management/3_4_Chatbot_Deletion.md)
- [3_5_Advanced_Configuration.md](./3.%20Chatbot%20Management/3_5_Advanced_Configuration.md)

### 4. Source Management
- [4_1_URL_Source_Creation.md](./4.%20Source%20Management/4_1_URL_Source_Creation.md)
- [4_2_File_Source_Creation.md](./4.%20Source%20Management/4_2_File_Source_Creation.md)
- [4_3_Source_Processing.md](./4.%20Source%20Management/4_3_Source_Processing.md)
- [4_4_Source_Refresh.md](./4.%20Source%20Management/4_4_Source_Refresh.md)
- [4_5_Discovery_Mode.md](./4.%20Source%20Management/4_5_Discovery_Mode.md)

### 5. Chat Functionality
- [5_1_Chat_Message_Flow.md](./5.%20Chat%20Functionality/5_1_Chat_Message_Flow.md)
- [5_2_RAG.md](./5.%20Chat%20Functionality/5_2_RAG.md)
- [5_3_Conversations_Feedback.md](./5.%20Chat%20Functionality/5_3_Conversations_Feedback.md)
- [5_4_Public_Widget.md](./5.%20Chat%20Functionality/5_4_Public_Widget.md)

### 6. Actions (Chatbot Actions)
- [6_1_Action_CRUD.md](./6.%20Actions%20%28Chatbot%20Actions%29/6_1_Action_CRUD.md)
- [6_2_Action_Execution.md](./6.%20Actions%20%28Chatbot%20Actions%29/6_2_Action_Execution.md)

### 7. Analytics
- [7_1_Analytics_Overview.md](./7.%20Analytics/7_1_Analytics_Overview.md)
- [7_2_Analytics_Trends.md](./7.%20Analytics/7_2_Analytics_Trends.md)

### 8. Organization & Workspace Management
- [8_1_Organization_Management.md](./8.%20Organization%20%26%20Workspace%20Management/8_1_Organization_Management.md)
- [8_2_Workspace_Management.md](./8.%20Organization%20%26%20Workspace%20Management/8_2_Workspace_Management.md)

### 9. Security & Validation
- [9_1_Input_Validation.md](./9.%20Security%20%26%20Validation/9_1_Input_Validation.md)
- [9_2_Auth_Authorization.md](./9.%20Security%20%26%20Validation/9_2_Auth_Authorization.md)

### 10. Database & Migrations
- [10_1_Migration_Integrity.md](./10.%20Database%20%26%20Migrations/10_1_Migration_Integrity.md)

### 11. External Integrations
- [11_1_OpenRouter_Integration.md](./11.%20External%20Integrations/11_1_OpenRouter_Integration.md)
- [11_2_Qdrant_Integration.md](./11.%20External%20Integrations/11_2_Qdrant_Integration.md)

### 12. Frontend UI Testing
- [12_1_Auth_Pages.md](./12.%20Frontend%20UI%20Testing/12_1_Auth_Pages.md)
- [12_2_Dashboard.md](./12.%20Frontend%20UI%20Testing/12_2_Dashboard.md)

### 13. Chat Widget Testing
- [13_1_Widget_Initialization.md](./13.%20Chat%20Widget%20Testing/13_1_Widget_Initialization.md)
- [13_2_Widget_Chat_Flow.md](./13.%20Chat%20Widget%20Testing/13_2_Widget_Chat_Flow.md)

### 14. Edge Cases & Error Handling
- [14_1_Edge_Cases.md](./14.%20Edge%20Cases%20%26%20Error%20Handling/14_1_Edge_Cases.md)

### 15. Localization & Internationalization
- [15_1_Localization.md](./15.%20Localization%20%26%20Internationalization/15_1_Localization.md)

### 16. Monitoring & Logging
- [16_1_Monitoring_Logging.md](./16.%20Monitoring%20%26%20Logging/16_1_Monitoring_Logging.md)

### 17. Deployment & Environment
- [17_1_Deployment.md](./17.%20Deployment%20%26%20Environment/17_1_Deployment.md)

### 18. Performance Testing
- [18_1_Performance.md](./18.%20Performance%20Testing/18_1_Performance.md)

### 19. Disaster Recovery & Backups
- [19_1_Disaster_Recovery.md](./19.%20Disaster%20Recovery%20%26%20Backups/19_1_Disaster_Recovery.md)

### 20. Compliance & Legal
- [20_1_Compliance.md](./20.%20Compliance%20%26%20Legal/20_1_Compliance.md)

### Final Pre-Production Checklist
- [Final_Checklist.md](./Final%20Pre-Production%20Checklist/Final_Checklist.md)

---

## Quick Start

### Run Backend Tests
```bash
cd /Users/onur/Documents/workspace/botla-co
make test
```

### Run Frontend Tests
```bash
cd frontend
npm run test
```

### Run E2E Tests
```bash
cd frontend
npm run test:e2e
```

---

## Test Coverage Summary

| Category | Files | Test Cases |
|----------|-------|------------|
| Authentication | 4 | ~30 |
| Plan Management | 5 | ~50 |
| Chatbot Management | 5 | ~40 |
| Source Management | 5 | ~45 |
| Chat Functionality | 4 | ~35 |
| Actions | 2 | ~20 |
| Analytics | 2 | ~12 |
| Organization | 2 | ~20 |
| Security | 2 | ~15 |
| Database | 1 | ~5 |
| Integrations | 2 | ~15 |
| Frontend | 2 | ~12 |
| Widget | 2 | ~12 |
| Edge Cases | 1 | ~15 |
| Other | 6 | ~20 |
| **Total** | **47** | **~350** |

---

## Priority Legend

- **Critical** - Must pass before any release
- **High** - Should pass before release
- **Medium** - Important for quality
- **Low** - Nice to have
