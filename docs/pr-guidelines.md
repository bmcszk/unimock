# PR Guidelines: Comprehensive Comment Tracking and Resolution

## 🚨 CRITICAL: Zero Tolerance for Missing Comments

This document provides **MANDATORY** guidelines for systematically tracking and addressing **ALL** PR review comments without exception. Missing even a single comment is unacceptable and undermines code quality.

## 📋 PR Workflow Overview

### Phase 1: Pre-PR Preparation
1. **Create feature branch** following Git Flow (NEVER commit to main/master)
2. **Implement with TDD** - Red-Green-Refactor cycle
3. **Run `make check`** after EVERY code change - NO EXCEPTIONS
4. **Ensure zero lint issues** before creating PR

### Phase 2: PR Creation
```bash
# Create PR with detailed description
gh pr create --title "Brief descriptive title" --body "## Summary\n## Changes\n## Testing\n## Checklist"

# Verify PR status
gh pr status
```

### Phase 3: Comment Tracking Protocol
🚨 **THIS IS THE MOST CRITICAL PHASE** - Follow EXACTLY to avoid missing comments

## 🔍 Step-by-Step Comment Retrieval Protocol

### 1. Retrieve ALL Reviews (Non-Negotiable)
```bash
# Get ALL review IDs first
gh api repos/:owner/:repo/pulls/:pr_number/reviews --jq '.[].id'

# Store ALL review IDs in variable
REVIEW_IDS=$(gh api repos/:owner/:repo/pulls/:pr_number/reviews --jq '.[].id' | tr '\n' ' ')

echo "Found reviews: $REVIEW_IDS"
```

### 2. Retrieve ALL Comments from ALL Reviews
```bash
# Extract comments from EACH review systematically
for review_id in $REVIEW_IDS; do
    echo "=== Review $review_id ==="
    gh api repos/:owner/:repo/pulls/:pr_number/reviews/$review_id/comments --jq '.[].id'
done
```

### 3. Get General PR Comments
```bash
# Also check for general PR comments (not part of reviews)
gh api repos/:owner/:repo/pulls/:pr_number/comments --jq '.[].id'
```

### 4. Create Comprehensive Tracking Document
```bash
# Create tracking document with ALL comment IDs
TOTAL_COMMENTS=$(gh api repos/:owner/:repo/pulls/:pr_number/comments --jq 'length')
echo "Total comments: $TOTAL_COMMENTS"

# Document must include: ID, file, line, issue description, priority, status
```

## 📝 Comment Categorization System

### Priority Classification:
- **HIGH (P0)**: Security vulnerabilities, architectural violations
- **MEDIUM (P1)**: Code quality issues, test failures, breaking changes
- **LOW (P2)**: Code style, documentation, nitpicks

### Status Tracking:
- **UNRESOLVED**: Initial state
- **IN_PROGRESS**: Being addressed
- **RESOLVED**: Fixed and verified
- **DISAGREED**: Documented disagreement with reasoning

## 🎯 Systematic Resolution Process

### For Each Comment:
1. **Read and understand** the issue completely
2. **Categorize** by priority (HIGH/MEDIUM/LOW)
3. **Implement fix** (if applicable)
4. **Verify fix** with `make check`
5. **Update tracking document** with resolution details
6. **Mark as resolved** in GitHub (if possible)

### Resolution Commands:
```bash
# For general comments that can be updated
gh api repos/:owner/:repo/pulls/:pr_number/comments/:comment_id \
  --method PATCH \
  --field body="✅ **RESOLVED**: [Detailed resolution explanation]

# For review comments - create new review to address
gh pr review --body "## Addressed Comments\n- ✅ [Comment ID]: [Resolution]"
```

## ⚠️ Common Pitfalls to AVOID

### ❌ NEVER DO THESE:
1. **Assume you got all comments** - ALWAYS verify complete retrieval
2. **Only check first few reviews** - REVIEWS can span multiple pages
3. **Skip general PR comments** - These are separate from review comments
4. **Ignore comments from automated tools** (Copilot, etc.)
5. **Mark as resolved without implementation** - Actually fix the issues
6. **Update tracking document last** - Track progress in real-time

### ✅ ALWAYS DO THESE:
1. **Get ALL review IDs first** before retrieving comments
2. **Cross-reference counts** - Verify total numbers match expectations
3. **Document EVERY comment** regardless of priority
4. **Use both weapons** - `make check` AND `make test-all`
5. **Commit after resolution** - Create clear, atomic commits
6. **Update documentation** - Track learnings for future PRs

## 🔧 GitHub CLI Command Reference

### Essential Commands:
```bash
# View PR with all comments
gh pr view --comments

# Get all reviews
gh api repos/:owner/:repo/pulls/:pr_number/reviews

# Get comments from specific review
gh api repos/:owner/:repo/pulls/:pr_number/reviews/:review_id/comments

# Get general PR comments
gh api repos/:owner/:repo/pulls/:pr_number/comments

# Create review to address multiple comments
gh pr review --body "## Resolutions\n- ✅ Comment 1: Fixed X\n- ✅ Comment 2: Updated Y"

# Add general comment
gh pr comment --body "Updated tracking document with all resolutions"
```

### Verification Commands:
```bash
# Count total comments across all sources
TOTAL_REVIEW_COMMENTS=$(gh api repos/:owner/:repo/pulls/:pr_number/reviews --jq '[.[].comments | length] | add')
TOTAL_GENERAL_COMMENTS=$(gh api repos/:owner/:repo/pulls/:pr_number/comments --jq 'length')
TOTAL_ALL=$((TOTAL_REVIEW_COMMENTS + TOTAL_GENERAL_COMMENTS))

echo "Total comments to address: $TOTAL_ALL"
```

## 📊 Tracking Document Template

```markdown
# PR Comment Tracking Document

**Pull Request:** #[number] - [brief title]
**Branch:** [feature-branch-name]
**Date:** [YYYY-MM-DD]
**Total Comments:** [count]

## Summary
This document tracks ALL unresolved PR review comments from the latest PR. Comments are categorized by priority and type for systematic resolution.

## High Priority Issues (Security/Architecture)
### 1. [Issue Title]
- **ID:** [comment_id]
- **File:** [path/to/file.ext]
- **Issue:** [detailed description]
- **Priority:** HIGH
- **Status:** [UNRESOLVED/IN_PROGRESS/RESOLVED]

## Medium Priority Issues (Code Quality)
[Same format as above]

## Low Priority Issues (Nitpicks)
[Same format as above]

## Task List Completion Status
### Phase 1: [Category] ([Status])
1. ✅ [Task description] - [Resolution details]
2. ⏳ [Task description] - [In progress]

## Success Criteria
- [ ] All security vulnerabilities addressed
- [ ] All architectural violations resolved
- [ ] All tests pass with `make check`
- [ ] Zero lint issues remain
- [ ] All comments documented and resolved
```

## 🚀 Quality Gates

### Before PR Merge:
- [ ] **All comments addressed** - ZERO exceptions
- [ ] **`make check` passes** - 216 tests, 0 lint issues
- [ ] **`make test-all` passes** - Complete test suite
- [ ] **Tracking document complete** - All comments documented
- [ ] **Code committed** - Atomic commits with clear messages

### Post-Merge:
- [ ] **Archive tracking document** - Save for future reference
- [ ] **Update guidelines** - Incorporate lessons learned
- [ ] **Team retrospective** - Discuss process improvements

## 🔄 Continuous Improvement

### Root Cause Analysis:
When comments are missed:
1. **Document the gap** in the retrieval process
2. **Update protocol** to prevent recurrence
3. **Train team members** on new procedures
4. **Audit PR process** regularly

### Metrics to Track:
- **Comment retrieval accuracy** (Target: 100%)
- **Time to resolution** (Target: <24 hours)
- **Quality gate failures** (Target: 0)
- **Process compliance** (Target: 100%)

---

## 🚨 FINAL REMINDER

**Missing PR comments is a SERIOUS process failure.** It indicates gaps in systematic approach and undermines code quality.

**ALWAYS:**
- Retrieve comments from ALL reviews
- Cross-verify total counts
- Document EVERY comment
- Implement fixes before marking resolved
- Use both quality weapons (`make check` + `make test-all`)

**NEVER:**
- Assume you got all comments without verification
- Skip reviews due to complexity
- Mark comments resolved without implementation
- Ignore comments from automated tools

**Zero tolerance means zero missed comments. No exceptions.**