#include <stdio.h>
#include <stdlib.h>

#define VM_MEM_SIZE 1000
#define VM_STACK_SIZE 1000
#define VM_RSTACK_SIZE 1000

static long mem[VM_MEM_SIZE];
static long stack[VM_STACK_SIZE];
static long rstack[VM_RSTACK_SIZE];
static long n = -1;
static long rn = -1;

// #ifndef inline
//   #define inline __inline__ __attribute__((always_inline))
// #endif

#define myerror(txt) \
  do { \
    printf("ERROR: " txt "\n"); \
    exit(0); \
  } while(0)

#ifndef DEBUG
  #define DEBUG 0
#endif

static inline void fvm_push(long i) {
  stack[++n] = i;
}

static inline long fvm_pop(void) {
#if DEBUG
  if (n < 0) myerror("D-Stack is empty in fvm_pop()");
#endif
  return stack[n--];
}

static inline void fvm_rpush(long i) {
  rstack[++rn] = i;
}

static inline long fvm_rpop(void) {
#if DEBUG
  if (rn < 0) myerror("R-Stack is empty in fvm_rpop()");
#endif
  return rstack[rn--];
}

static inline double fvm_fpop(void) {
  long v;
  v = fvm_pop();
  return *(double*)&v;
}

static inline void fvm_fpush(double i) {
  long v;
  v = *((long*)&i);
  fvm_push(v);
}

static inline void fvm_lv(void) {
  fvm_push(mem[fvm_pop()]);
}

static inline void fvm_lsi(void) {
  long a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(a > b);
}

static inline void fvm_gri(void) {
  long a, b;
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
  long a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(b - a);
}

static inline void fvm_dvi(void) {
  long a, b;
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
  double a, b;
  a = fvm_fpop();
  b = fvm_fpop();
  fvm_fpush(b - a);
}

static inline void fvm_dvf(void) {
  double a, b;
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
  double a, b;
  a = fvm_fpop();
  b = fvm_fpop();
  fvm_push(a > b);
}

static inline void fvm_grf(void) {
  double a, b;
  a = fvm_fpop();
  b = fvm_fpop();
  fvm_push(a < b);
}

static inline void fvm_pra(void) {
  printf("%c", (int)fvm_pop());
}

static inline void fvm_rdi(void) {
  long i;
  scanf("%ld", &i);
  fvm_push(i);
}

static inline void fvm_eqi(void) {
  long a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(a == b);
}

static inline void fvm_xor(void) {
  long a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(a ^ b);
}

static inline void fvm_and(void) {
  long a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(a && b);
}

static inline void fvm_or(void) {
  long a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(a || b);
}

static inline void fvm_not(void) {
  fvm_push(fvm_pop() == 0);
}

static inline void fvm_str(void) {
  long a, b;
  a = fvm_pop();
  b = fvm_pop();
  mem[a] = b;
}

static inline void fvm_dup(void) {
  long value;
  value = fvm_pop();
  fvm_push(value);
  fvm_push(value);
}

static inline void fvm_pck(void) {
  long v;
  v = fvm_pop();
  fvm_push(stack[n-v]);
}

static inline void fvm_ovr(void) {
  long a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push(b);
  fvm_push(a);
  fvm_push(b);
}

static inline void fvm_tvr(void) {
  long a, b, c, d;
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
  long a, b, c, d;
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
  long a;
  a = fvm_pop();
  fvm_push(a);

  if (a != 0) {
    fvm_push(a);
  }
}

static inline void fvm_rot(void) {
  long a, b, c;
  a = fvm_pop();
  b = fvm_pop();
  c = fvm_pop();
  fvm_push(b);
  fvm_push(a);
  fvm_push(c);
}

static inline void fvm_nrt(void) {
  long a, b, c;
  a = fvm_pop();
  b = fvm_pop();
  c = fvm_pop();
  fvm_push(a);
  fvm_push(c);
  fvm_push(b);
}

static inline void fvm_tdp(void) {
  long a, b;
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
  long a, b;
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
  fvm_push(rstack[rn]);
}

static inline void fvm_ttr(void) {
  long a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_rpush(b);
  fvm_rpush(a);
}

static inline void fvm_tfr(void) {
  long a, b;
  a = fvm_rpop();
  b = fvm_rpop();
  fvm_push(b);
  fvm_push(a);
}

static inline void fvm_trf(void) {
  fvm_push(rstack[rn-1]);
  fvm_push(rstack[rn]);
}

static inline void fvm_sys(void) {
  long sys, value;
  double dvalue;

  sys = fvm_pop();

  switch (sys) {
  case 3:
    // i>f
    value = fvm_pop();
    fvm_fpush((double)value);
    break;
  case 4:
    // f>i
    dvalue = fvm_fpop();
    fvm_push((long)dvalue);
    break;
  case 10:
    // allocate
    // TODO: not implemented
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
  fvm_push((long)f);
}

static inline void fvm_exc(void) {
  void (*f)(void);

  f = (void (*)(void))fvm_pop();
  (*f)();
}

#define fvm_stp() \
  do { \
    printf("\n"); \
    return 0; \
  } while (0)
