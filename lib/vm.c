#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <stddef.h>
#include <time.h>

#ifndef DEBUG
  #define DEBUG 0
#endif

#define VM_MEM_SIZE 1000
#define VM_STACK_SIZE 1000
#define VM_RSTACK_SIZE 1000

typedef int64_t cell_t;
typedef double  fcell_t;

static cell_t fvm_mem[VM_MEM_SIZE];
static cell_t fvm_stack[VM_STACK_SIZE];
static cell_t fvm_rstack[VM_RSTACK_SIZE];
static ptrdiff_t fvm_n = -1;
static ptrdiff_t fvm_rn = -1;

#if DEBUG
static cell_t fvm_nmax = 0;
static cell_t fvm_rmax = 0;
#endif

static clock_t fvm_begin = 0;

// #ifndef inline
//   #define inline __inline__ __attribute__((always_inline))
// #endif

#define myerror(txt) \
  do { \
    printf("ERROR: " txt "\n"); \
    exit(0); \
  } while(0)

static inline void fvm_push(cell_t i) {
  fvm_stack[++fvm_n] = i;
#if DEBUG
  if (fvm_n > fvm_nmax) {
    fvm_nmax = fvm_n;
  }
#endif
}

static inline cell_t fvm_pop(void) {
#if DEBUG
  if (fvm_n < 0) myerror("fvm_stack is empty in fvm_pop()");
#endif
  return fvm_stack[fvm_n--];
}

static inline void fvm_rpush(cell_t i) {
  fvm_rstack[++fvm_rn] = i;
#if DEBUG
  if (fvm_rn > fvm_rmax) {
    fvm_rmax = fvm_rn;
  }
#endif
}

static inline cell_t fvm_rpop(void) {
#if DEBUG
  if (fvm_rn < 0) myerror("fvm_rstack is empty in fvm_rpop()");
#endif
  return fvm_rstack[fvm_rn--];
}

static inline fcell_t fvm_fpop(void) {
  cell_t v;
  v = fvm_pop();
  return *(fcell_t*)&v;
}

static inline void fvm_fpush(fcell_t i) {
  cell_t v;
  v = *((cell_t*)&i);
  fvm_push(v);
}

static inline void fvm_lv(void) {
  fvm_push(fvm_mem[fvm_pop()]);
}

static inline void fvm_lsi(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(a > b);
}

static inline void fvm_gri(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(a < b);
}

static inline int fvm_jin(void) {
  return fvm_pop() == 0;
}

static inline void fvm_adi(void) {
  fvm_push(fvm_pop() + fvm_pop());
}

static inline void fvm_sbi(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(b - a);
}

static inline void fvm_dvi(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(b / a);
}

static inline void fvm_mli(void) {
  fvm_push(fvm_pop() * fvm_pop());
}

static inline void fvm_adf(void) {
  fvm_fpush(fvm_fpop() + fvm_fpop());
}

static inline void fvm_sbf(void) {
  fcell_t a, b;
  a = fvm_fpop();
  b = fvm_fpop();
  fvm_fpush(b - a);
}

static inline void fvm_dvf(void) {
  fcell_t a, b;
  a = fvm_fpop();
  b = fvm_fpop();
  fvm_fpush(b / a);
}

static inline void fvm_mlf(void) {
  fvm_fpush(fvm_fpop() * fvm_fpop());
}

static inline void fvm_pri(void) {
  printf("%ld", fvm_pop());
}

static inline void fvm_prf(void) {
  printf("%f", fvm_fpop());
}

static inline void fvm_lsf(void) {
  fcell_t a, b;
  a = fvm_fpop();
  b = fvm_fpop();
  fvm_push(a > b);
}

static inline void fvm_grf(void) {
  fcell_t a, b;
  a = fvm_fpop();
  b = fvm_fpop();
  fvm_push(a < b);
}

static inline void fvm_pra(void) {
  printf("%c", (int)fvm_pop());
}

static inline void fvm_rdi(void) {
  cell_t i;
  scanf("%ld", &i);
  fvm_push(i);
}

static inline void fvm_eqi(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(a == b);
}

static inline void fvm_xor(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(a ^ b);
}

static inline void fvm_and(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(a && b);
}

static inline void fvm_or(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(a || b);
}

static inline void fvm_not(void) {
  fvm_push(fvm_pop() == 0);
}

static inline void fvm_str(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_mem[a] = b;
}

static inline void fvm_dup(void) {
  cell_t value;
  value = fvm_pop();
  fvm_push(value);
  fvm_push(value);
}

static inline void fvm_pck(void) {
  cell_t v;
  v = fvm_pop();
  fvm_push(fvm_stack[fvm_n-v]);
}

static inline void fvm_ovr(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(b);
  fvm_push(a);
  fvm_push(b);
}

static inline void fvm_tvr(void) {
  cell_t a, b, c, d;
  a = fvm_pop();
  b = fvm_pop();
  c = fvm_pop();
  d = fvm_pop();
  fvm_push(d);
  fvm_push(c);
  fvm_push(b);
  fvm_push(a);
  fvm_push(d);
  fvm_push(c);
}

static inline void fvm_twp(void) {
  cell_t a, b, c, d;
  a = fvm_pop();
  b = fvm_pop();
  c = fvm_pop();
  d = fvm_pop();
  fvm_push(b);
  fvm_push(a);
  fvm_push(d);
  fvm_push(c);
}

static inline void fvm_qdp(void) {
  cell_t a;
  a = fvm_pop();
  fvm_push(a);

  if (a != 0) {
    fvm_push(a);
  }
}

static inline void fvm_rot(void) {
  cell_t a, b, c;
  a = fvm_pop();
  b = fvm_pop();
  c = fvm_pop();
  fvm_push(b);
  fvm_push(a);
  fvm_push(c);
}

static inline void fvm_nrt(void) {
  cell_t a, b, c;
  a = fvm_pop();
  b = fvm_pop();
  c = fvm_pop();
  fvm_push(a);
  fvm_push(c);
  fvm_push(b);
}

static inline void fvm_tdp(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(b);
  fvm_push(a);
  fvm_push(b);
  fvm_push(a);
}

static inline void fvm_drp(void) {
  fvm_pop();
}

static inline void fvm_swp(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(a);
  fvm_push(b);
}

static inline void fvm_tr(void) {
  fvm_rpush(fvm_pop());
}

static inline void fvm_fr(void) {
  fvm_push(fvm_rpop());
}

static inline void fvm_rf(void) {
  fvm_push(fvm_rstack[fvm_rn]);
}

static inline void fvm_ttr(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_rpush(b);
  fvm_rpush(a);
}

static inline void fvm_tfr(void) {
  cell_t a, b;
  a = fvm_rpop();
  b = fvm_rpop();
  fvm_push(b);
  fvm_push(a);
}

static inline void fvm_trf(void) {
  fvm_push(fvm_rstack[fvm_rn-1]);
  fvm_push(fvm_rstack[fvm_rn]);
}

static inline void fvm_sys(void) {
  cell_t sys, value;
  fcell_t dvalue;

  sys = fvm_pop();

  switch (sys) {
  case 3:
    // i>f
    value = fvm_pop();
    fvm_fpush((fcell_t)value);
    break;
  case 4:
    // f>i
    dvalue = fvm_fpop();
    fvm_push((cell_t)dvalue);
    break;
  case 10:
    // allocate
    fvm_pop(); // pop argument
    break;
  case 11:
    // memsize
    fvm_push(VM_MEM_SIZE);
    break;
  default:
    printf("ERROR: Unknown sys command\n");
    break;
  }
}

static inline void fvm_ref(void (*f)(void)) {
  fvm_push((cell_t)f);
}

static inline void fvm_exc(void) {
  void (*f)(void);

  f = (void (*)(void))fvm_pop();
  (*f)();
}

static inline void fvm_time(void) {
  fvm_begin = clock();
}

static inline void fvm_stp(void) {
#if DEBUG
  printf("\nstack max: %ld rstack max: %ld\n", fvm_nmax, fvm_rmax);
#endif
  if (fvm_begin != 0) {
    clock_t fvm_end = clock();
    double time_spend = (double)(fvm_end - fvm_begin);
    printf("\ntime: %fs\n", time_spend / CLOCKS_PER_SEC);
  }
  exit(0);
}
