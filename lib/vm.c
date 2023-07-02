#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <stddef.h>
#include <string.h>
#include <unistd.h>
#include <time.h>

#ifndef DEBUG
  #define DEBUG 0
#endif

#define VM_STACK_SIZE 200
#define VM_RSTACK_SIZE 50

typedef union u_cell {
  int64_t value;
  double  dvalue;
  void    (*func)(void);
} cell_t;

static int64_t fvm_argc = 0;
static char** fvm_argv = NULL;
static int64_t fvm_mem_size = 0;
static cell_t* fvm_mem = NULL;
static cell_t fvm_stack[VM_STACK_SIZE];
static cell_t fvm_rstack[VM_RSTACK_SIZE];
static ptrdiff_t fvm_n = -1;
static ptrdiff_t fvm_rn = -1;

#if DEBUG
static int64_t fvm_nmax = 0;
static int64_t fvm_rmax = 0;
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

static inline cell_t fvm_cell(int64_t i) {
  return (cell_t){ .value = i };
}

static inline cell_t fvm_cell_d(double i) {
  return (cell_t){ .dvalue = i };
}

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

static inline void fvm_lv(void) {
  fvm_push(fvm_mem[fvm_pop().value]);
}

static inline void fvm_lsi(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push((cell_t){ .value = a.value > b.value });
}

static inline void fvm_gri(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push((cell_t){ .value = a.value < b.value });
}

static inline int fvm_jin(void) {
  return fvm_pop().value == 0;
}

static inline void fvm_adi(void) {
  fvm_push((cell_t){ .value = fvm_pop().value + fvm_pop().value });
}

static inline void fvm_sbi(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push((cell_t){ .value = b.value - a.value });
}

static inline void fvm_dvi(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push((cell_t){ .value = b.value / a.value });
}

static inline void fvm_mli(void) {
  fvm_push((cell_t){ .value = fvm_pop().value * fvm_pop().value });
}

static inline void fvm_adf(void) {
  fvm_push((cell_t){ .dvalue = fvm_pop().dvalue + fvm_pop().dvalue });
}

static inline void fvm_sbf(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push((cell_t){ .dvalue = b.dvalue - a.dvalue });
}

static inline void fvm_dvf(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push((cell_t){ .dvalue = b.dvalue / a.dvalue });
}

static inline void fvm_mlf(void) {
  fvm_push((cell_t){ .dvalue = fvm_pop().dvalue * fvm_pop().dvalue });
}

static inline void fvm_pri(void) {
  printf("%ld", fvm_pop().value);
}

static inline void fvm_prf(void) {
  printf("%f", fvm_pop().dvalue);
}

static inline void fvm_lsf(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push((cell_t){ .value = a.dvalue > b.dvalue });
}

static inline void fvm_grf(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push((cell_t){ .value = a.dvalue < b.dvalue });
}

static inline void fvm_pra(void) {
  printf("%c", (int)(fvm_pop().value));
}

static inline void fvm_rdi(void) {
  cell_t i;
  scanf("%ld", &(i.value));
  fvm_push(i);
}

static inline void fvm_eqi(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push((cell_t){ .value = a.value == b.value });
}

static inline void fvm_xor(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push((cell_t){ .value = a.value ^ b.value });
}

static inline void fvm_and(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push((cell_t){ .value = a.value && b.value });
}

static inline void fvm_or(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_push((cell_t){ .value = a.value || b.value });
}

static inline void fvm_not(void) {
  fvm_push((cell_t){ .value = fvm_pop().value == 0 });
}

static inline void fvm_str(void) {
  cell_t a, b;
  a = fvm_pop();
  b = fvm_pop();
  fvm_mem[a.value] = b;
}

static inline void fvm_dup(void) {
  cell_t a;
  a = fvm_pop();
  fvm_push(a);
  fvm_push(a);
}

static inline void fvm_pck(void) {
  cell_t v;
  v = fvm_pop();
  fvm_push(fvm_stack[fvm_n-v.value]);
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

  if (a.value != 0) {
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

static inline char* fvm_getstring() {
  cell_t str = fvm_pop();
  cell_t len = fvm_mem[str.value];
  char *buffer = (char*)malloc(len.value+1);

  if (buffer == NULL) {
    printf("ERROR: Unable to allocate memory\n");
    exit(0);
  }

  for (int64_t i = 0; i < len.value; i++) {
    buffer[i] = (char)fvm_mem[str.value+1+i].value;
  }

  buffer[len.value] = '\0';

  return buffer;
}

static inline void fvm_setstring(const char* str) {
  cell_t addr = fvm_pop();
  size_t len = strlen(str);

  fvm_mem[addr.value].value = (int64_t)len;

  for (size_t i = 0; i < len; i++) {
    fvm_mem[addr.value+1+i].value = (int64_t)str[i];
  }
}

static inline void fvm_free(void) {
  if (fvm_mem_size > 0) {
    free(fvm_mem);
    fvm_mem = NULL;
    fvm_mem_size = 0;
  }
}

static inline void fvm_copy(cell_t *dest, int64_t dest_size, const cell_t *src, int64_t src_size) {
  int64_t n = dest_size < src_size ? dest_size : src_size;
  if (n == 0) return;
  memcpy(dest, src, sizeof(cell_t) * n);
}

static inline void fvm_sys(void) {
  cell_t sys = fvm_pop();

  switch (sys.value) {
  case 3:
    // i>f
    {
      cell_t c;
      c.dvalue = fvm_pop().value;
      fvm_push(c);
    }
    break;
  case 4:
    // f>i
    {
      cell_t c;
      c.value = fvm_pop().dvalue;
      fvm_push(c);
    }
    break;
  case 10:
    // allocate
    {
      cell_t n = fvm_pop();
      if (n.value == 0) {
        fvm_free();
      } else {
        cell_t *tmp = (cell_t*)calloc((size_t)n.value, sizeof(cell_t));
        if (tmp == NULL) {
          printf("ERROR: Unable to allocate memory\n");
          exit(0);
        }
        fvm_copy(tmp, n.value, fvm_mem, fvm_mem_size);
        fvm_free();
        fvm_mem_size = n.value;
        fvm_mem = tmp;
      }
    }
    break;
  case 11:
    // memsize
    fvm_push(fvm_cell(fvm_mem_size));
    break;
  case 12:
    // compare
    {
      char *str1 = fvm_getstring();
      char *str2 = fvm_getstring();

      if (strcmp(str1, str2) == 0) {
        fvm_push(fvm_cell(1));
      } else {
        fvm_push(fvm_cell(0));
      }

      free(str1);
      free(str2);
    }
    break;
  case 13:
    // shell
    {
      char *buffer = fvm_getstring();
      system(buffer);
      free(buffer);
    }
    break;
  case 14:
    // system
    {
      char *buffer = fvm_getstring();
      execl(buffer, buffer, (char*)NULL);
      free(buffer);
    }
    break;
  case 15:
    // file
    {
      char *buffer = fvm_getstring();
      if (access(buffer, F_OK) == 0) {
        fvm_push(fvm_cell(1));
      } else {
        fvm_push(fvm_cell(0));
      }
      free(buffer);
    }
    break;
  case 16:
    // argc
    fvm_push(fvm_cell(fvm_argc));
    break;
  case 17:
    // argv
    {
      cell_t n = fvm_pop();
      char *arg = fvm_argv[(size_t)n.value];
      fvm_setstring(arg);
    }
    break;
  default:
    printf("ERROR: Unknown sys command\n");
    break;
  }
}

static inline void fvm_ref(void (*f)(void)) {
  fvm_push((cell_t){ .func = f });
}

static inline void fvm_exc(void) {
  fvm_pop().func();
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

  fvm_free();
  exit(0);
}
