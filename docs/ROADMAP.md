# pyscn Development Roadmap

## 🎯 Current Sprint: MVP (August 2025)

### Week 1 (Aug 6-12) - Foundation
- [x] Project setup and repository initialization
- [ ] Tree-sitter Python integration
- [ ] Basic AST construction
- [ ] CFG algorithm implementation

### Week 2 (Aug 13-19) - Dead Code Detection
- [ ] CFG-based dead code detection
- [ ] Unreachable code detection
- [ ] Unused function/variable detection
- [ ] Cyclomatic complexity calculation

### Week 3 (Aug 20-26) - Clone Detection
- [ ] APTED algorithm implementation
- [ ] Structural clone detection
- [ ] Similarity threshold configuration
- [ ] Clone report generation

### Week 4 (Aug 27-Sep 2) - Polish & Release
- [ ] CLI implementation (Cobra)
- [ ] Output formatters (JSON/Text/SARIF)
- [ ] Test suite completion
- [ ] Documentation
- [ ] **Open Source Release** 🚀

## 📊 Progress Tracking

### Core Technologies
- [ ] **Tree-sitter Integration**
  - [ ] Python grammar setup
  - [ ] AST traversal
  - [ ] Node type mapping
  
- [ ] **CFG Implementation**
  - [ ] Basic block construction
  - [ ] Control flow edges
  - [ ] Reachability analysis
  
- [ ] **APTED Implementation**
  - [ ] Tree edit distance algorithm
  - [ ] Node comparison logic
  - [ ] Performance optimization

### Features Status
| Feature | Status | Priority | Target Date |
|---------|--------|----------|-------------|
| Python Parsing | 🔄 In Progress | P0 | Week 1 |
| Dead Code Detection | ⏳ Planned | P0 | Week 2 |
| Clone Detection | ⏳ Planned | P0 | Week 3 |
| CLI Interface | ⏳ Planned | P0 | Week 4 |
| Config Files | ⏳ Planned | P1 | Week 4 |
| VS Code Extension | 🔮 Future | P2 | v0.2.0 |
| LLM Integration | 🔮 Future | P3 | v0.3.0 |

## 🏆 Milestones

### v0.1.0 - MVP Release (Sep 6, 2025)
- ✅ Basic Python analysis with Tree-sitter
- ✅ CFG-based dead code detection
- ✅ APTED-based clone detection
- ✅ CLI with basic commands
- ✅ Configuration file support

### v0.2.0 - Extended Features (Oct 2025)
- ⏳ Dependency analysis
- ⏳ Performance pattern detection
- ⏳ VS Code extension (basic)
- ⏳ GitHub Actions integration

### v0.3.0 - Pro Features (Nov 2025)
- ⏳ LLM integration for suggestions
- ⏳ Advanced refactoring proposals
- ⏳ Team collaboration features
- ⏳ Web dashboard

## 📈 Success Metrics

### Technical KPIs
- [ ] Analysis speed: >10,000 lines/second
- [ ] False positive rate: <5%
- [ ] Memory usage: <100MB for 100k LOC

### Community KPIs
- [ ] 100 GitHub stars (Month 1)
- [ ] 10 contributors (Month 2)
- [ ] 1000 downloads (Month 3)

## 🤝 How to Contribute

1. Pick an unassigned task from the current sprint
2. Create an issue to track your work
3. Submit a PR following our guidelines
4. Help review other PRs

## 📝 Notes

- All dates are tentative and may adjust based on progress
- Community feedback will influence feature prioritization
- Performance optimization is ongoing throughout all phases